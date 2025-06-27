// This module encapsulates all chat, audio, and streaming logic.
export function initChatModule(coachAgentURL) {
    // --- DOM Elements ---
    const chatHistory = document.getElementById('chat-history');
    const chatMessageInput = document.getElementById('chat-message-input');
    const chatSendButton = document.getElementById('chat-send-button');
    const chatDeleteSessionButton = document.getElementById('chat-delete-session-button');
    const vuMeterLevel = document.getElementById('vu-meter-level');
    const micGainSlider = document.getElementById('mic-gain-slider');
    const streamToggleSwitch = document.getElementById('stream-toggle-switch');
    const showTextToggleSwitch = document.getElementById('show-text-toggle-switch');
    const playAudioToggleSwitch = document.getElementById('play-audio-toggle-switch');
    const debugModeToggleSwitch = document.getElementById('debug-mode-toggle-switch');
    const micMuteToggleSwitch = document.getElementById('mic-mute-toggle-switch');


    // --- State Variables ---
    let audioContext;
    let microphoneStream;
    let micGainNode;
    let workletNode;
    let audioStreamSocket;
    let isStreaming = false;
    let isMuted = true;
    let audioQueue = [];
    let isAudioPlaying = false;
    let currentAudioSource = null;
    let isTearingDown = false;
    let isConnecting = false;
    let currentAgentMessageElement = null;

    // --- Constants & Configuration ---
    const AGENT_AUDIO_SAMPLE_RATE = 24000;
    const TARGET_SAMPLE_RATE = 16000;

    // --- Audio Worklet ---
    const audioWorkletProcessor = `
        class AudioStreamerProcessor extends AudioWorkletProcessor {
            process(inputs, outputs, parameters) {
                const inputData = inputs[0][0];
                if (inputData) this.port.postMessage(inputData);
                return true;
            }
        }
        registerProcessor('audio-streamer-processor', AudioStreamerProcessor);
    `;
    const workletBlob = new Blob([audioWorkletProcessor], {type: 'application/javascript'});
    const workletURL = URL.createObjectURL(workletBlob);

    // --- UI & Helper Functions ---
    function addMessageToChat(text, type = 'user') {
        // If debug mode is off, don't show system-info messages
        if (type === 'system-info' && !debugModeToggleSwitch.checked) {
            return null;
        }

        const messageElement = document.createElement('div');
        messageElement.classList.add('chat-message', type);
        if (type === 'system-error') messageElement.style.color = 'red';
        if (type === 'system-success') messageElement.style.color = 'green';
        if (type === 'system-info') messageElement.style.color = 'blue';
        messageElement.textContent = text;
        chatHistory.appendChild(messageElement);
        chatHistory.scrollTop = chatHistory.scrollHeight;
        return messageElement;
    }

    function updateVuMeter(data) {
        if (!data || data.length === 0) {
            vuMeterLevel.style.width = '0%';
            return;
        }
        let sumSquares = 0.0;
        for (const sample of data) sumSquares += sample * sample;
        const rms = Math.sqrt(sumSquares / data.length);
        const volume = Math.min(100, rms * 200);
        vuMeterLevel.style.width = `${volume}%`;
    }

    function setControlsDisabled(disabled) {
        chatMessageInput.disabled = disabled;
        chatSendButton.disabled = disabled;
        micGainSlider.disabled = disabled;
        micMuteToggleSwitch.disabled = disabled;
        // showTextToggleSwitch and playAudioToggleSwitch are now always enabled.
    }

    // --- Audio Processing Functions (No changes here) ---
    function floatTo16BitPCM(input) {
        const output = new DataView(new ArrayBuffer(input.length * 2));
        for (let i = 0; i < input.length; i++) {
            const s = Math.max(-1, Math.min(1, input[i]));
            const pcmSample = s < 0 ? s * 0x8000 : s * 0x7FFF;
            output.setInt16(i * 2, pcmSample, true);
        }
        return output.buffer;
    }

    function arrayBufferToBase64(buffer) {
        let binary = '';
        const bytes = new Uint8Array(buffer);
        for (let i = 0; i < bytes.byteLength; i++) binary += String.fromCharCode(bytes[i]);
        return window.btoa(binary);
    }

    function base64ToArrayBuffer(base64) {
        const binary_string = window.atob(base64);
        const len = binary_string.length;
        const bytes = new Uint8Array(len);
        for (let i = 0; i < len; i++) bytes[i] = binary_string.charCodeAt(i);
        return bytes.buffer;
    }

    function pcm16ToFloat(arrayBuffer) {
        const int16Array = new Int16Array(arrayBuffer);
        const float32Array = new Float32Array(int16Array.length);
        for (let i = 0; i < int16Array.length; i++) float32Array[i] = int16Array[i] / 32768.0;
        return float32Array;
    }

    // --- Audio Playback Queue (No changes here) ---
    function stopAndClearAudioQueue() {
        console.log("Interrupting audio playback and clearing queue.");
        audioQueue = [];
        if (currentAudioSource) currentAudioSource.stop();
    }

    function playNextChunkInQueue() {
        if (audioQueue.length === 0) {
            isAudioPlaying = false;
            currentAudioSource = null;
            return;
        }
        isAudioPlaying = true;
        const float32Audio = audioQueue.shift();

        if (!audioContext || audioContext.state !== 'running') {
            console.warn("AudioContext not running, cannot play audio. Re-queuing chunk.");
            audioQueue.unshift(float32Audio);
            isAudioPlaying = false;
            return;
        }

        const audioBuffer = audioContext.createBuffer(1, float32Audio.length, AGENT_AUDIO_SAMPLE_RATE);
        audioBuffer.copyToChannel(float32Audio, 0);
        const source = audioContext.createBufferSource();
        currentAudioSource = source;
        source.buffer = audioBuffer;
        source.connect(audioContext.destination);
        source.onended = playNextChunkInQueue;
        source.start();
    }

    // --- Microphone & UI State ---
    function updateMicMuteToggleState() {
        if (!isStreaming) {
            micMuteToggleSwitch.checked = false;
            micMuteToggleSwitch.disabled = true;
        } else {
            micMuteToggleSwitch.disabled = false;
            micMuteToggleSwitch.checked = !isMuted;
        }
    }

    // --- MODIFIED: Added user-facing messages for mic state changes ---
    function toggleMute() {
        if (!isStreaming) {
            console.warn("Cannot toggle mute: not streaming.");
            return;
        }
        isMuted = !micMuteToggleSwitch.checked;
        console.log(`Microphone is now ${isMuted ? 'muted' : 'unmuted'}.`);

        if (isMuted) {
            addMessageToChat("Mic muted", "system-success");
        } else {
            addMessageToChat("Audio connected. You can now speak.", "system-success");
        }
    }

    // --- Core Streaming Logic ---
    function setupStreamingPipeline() {
        return new Promise((resolve, reject) => {
            if (isStreaming) {
                resolve();
                return;
            }
            audioQueue = [];
            isAudioPlaying = false;
            currentAudioSource = null;

            try {
                if (!audioContext || audioContext.state === 'closed') {
                    console.log("Creating new AudioContext.");
                    audioContext = new (window.AudioContext || window.webkitAudioContext)();
                }
                if (audioContext.state === 'suspended') {
                    audioContext.resume();
                }
                audioContext.audioWorklet.addModule(workletURL).catch(reject);
            } catch (e) {
                console.error("AudioContext/Worklet error:", e);
                addMessageToChat("The audio context could not be started.", "system-error");
                reject(e);
                return;
            }

            const wsURL = `${coachAgentURL.replace('http', 'ws')}/api/v1/chat/stream`;
            audioStreamSocket = new WebSocket(wsURL);

            audioStreamSocket.onopen = async () => {
                addMessageToChat("Streaming connection established.", "system-info");
                try {
                    microphoneStream = await navigator.mediaDevices.getUserMedia({
                        audio: {
                            sampleRate: audioContext.sampleRate,
                            channelCount: 1,
                            echoCancellation: true,
                            noiseSuppression: true
                        }
                    });
                    micGainNode = audioContext.createGain();
                    micGainNode.gain.value = parseFloat(micGainSlider.value);
                    const source = audioContext.createMediaStreamSource(microphoneStream);
                    source.connect(micGainNode);
                    workletNode = new AudioWorkletNode(audioContext, 'audio-streamer-processor');
                    micGainNode.connect(workletNode);

                    workletNode.port.onmessage = async (event) => {
                        if (isMuted || !audioStreamSocket || audioStreamSocket.readyState !== WebSocket.OPEN) {
                            updateVuMeter(null);
                            return;
                        }
                        const inputFloat32Array = event.data;
                        const originalBuffer = audioContext.createBuffer(1, inputFloat32Array.length, audioContext.sampleRate);
                        originalBuffer.copyToChannel(inputFloat32Array, 0);
                        const offlineContext = new OfflineAudioContext(1, originalBuffer.duration * TARGET_SAMPLE_RATE, TARGET_SAMPLE_RATE);
                        const bufferSource = offlineContext.createBufferSource();
                        bufferSource.buffer = originalBuffer;
                        bufferSource.connect(offlineContext.destination);
                        bufferSource.start();
                        const resampledBuffer = await offlineContext.startRendering();
                        const resampledData = resampledBuffer.getChannelData(0);
                        const pcm16Buffer = floatTo16BitPCM(resampledData);
                        const base64Data = arrayBufferToBase64(pcm16Buffer);
                        audioStreamSocket.send(JSON.stringify({mime_type: "audio/pcm", data: base64Data}));
                        updateVuMeter(inputFloat32Array);
                    };

                    isStreaming = true;
                    isMuted = !micMuteToggleSwitch.checked;
                    updateMicMuteToggleState();
                    resolve();
                } catch (err) {
                    addMessageToChat(`Microphone error: ${err.message}`, "system-error");
                    teardownStreamingPipeline();
                    reject(err);
                }
            };

            audioStreamSocket.onmessage = (event) => {
                if (typeof event.data !== 'string') return;

                try {
                    const message = JSON.parse(event.data);

                    if (showTextToggleSwitch.checked && message.mime_type === 'text/plain' && message.data) {
                        if (!currentAgentMessageElement) {
                            currentAgentMessageElement = addMessageToChat(message.data, 'agent');
                        } else {
                            currentAgentMessageElement.textContent += message.data;
                            chatHistory.scrollTop = chatHistory.scrollHeight;
                        }
                    }

                    if (playAudioToggleSwitch.checked && message.mime_type === 'audio/pcm' && message.data) {
                        const pcm16Buffer = base64ToArrayBuffer(message.data);
                        const float32Audio = pcm16ToFloat(pcm16Buffer);
                        if (float32Audio.length > 0 && !float32Audio.every(s => s === 0)) {
                            audioQueue.push(float32Audio);
                            if (!isAudioPlaying) playNextChunkInQueue();
                        }
                    }

                    if (message.turn_complete || message.interrupted) {
                        if (message.interrupted) {
                            addMessageToChat("Agent interrupted.", "system-success");
                            stopAndClearAudioQueue();
                        }
                        if (message.turn_complete) {
                            addMessageToChat("Agent turn complete.", "system-info");
                        }
                        currentAgentMessageElement = null;
                    }

                } catch (e) {
                    console.error("Error processing received message:", e, "Data:", event.data);
                    currentAgentMessageElement = null;
                }
            };

            // --- MODIFIED: Changed the unexpected close message to be less alarming and green ---
            audioStreamSocket.onclose = (event) => {
                const wasManualClose = event.reason === "Client initiated teardown";
                if (!wasManualClose) {
                    addMessageToChat("Streaming stopped. Please re-enable the connection.", "system-success");
                }
                teardownStreamingPipeline();
            };

            audioStreamSocket.onerror = (error) => {
                console.error("WebSocket error:", error);
                reject(new Error("WebSocket connection error."));
            };
        });
    }

    function teardownStreamingPipeline() {
        if (isTearingDown) return;
        isTearingDown = true;
        console.log("Tearing down streaming pipeline...");

        if (microphoneStream) {
            microphoneStream.getTracks().forEach(track => track.stop());
            microphoneStream = null;
        }
        if (workletNode) {
            workletNode.port.close();
            workletNode.disconnect();
            workletNode = null;
        }
        if (micGainNode) {
            micGainNode.disconnect();
            micGainNode = null;
        }
        if (audioStreamSocket) {
            if (audioStreamSocket.readyState === WebSocket.OPEN || audioStreamSocket.readyState === WebSocket.CONNECTING) {
                audioStreamSocket.close(1000, "Client initiated teardown");
            }
            audioStreamSocket = null;
        }

        if (audioContext && audioContext.state !== 'closed') {
            audioContext.close();
        }
        audioContext = null;

        stopAndClearAudioQueue();

        isStreaming = false;
        isMuted = true;
        currentAgentMessageElement = null;

        if (streamToggleSwitch.checked) {
            streamToggleSwitch.checked = false;
        }
        setControlsDisabled(true);
        updateMicMuteToggleState();
        updateVuMeter(null);

        console.log("Pipeline teardown complete.");
        isTearingDown = false;
    }

    function sendMessage() {
        const messageText = chatMessageInput.value.trim();
        if (messageText === '') return;
        if (!isStreaming) {
            addMessageToChat("Connection is not active.", "system-error");
            return;
        }
        if (currentAgentMessageElement) {
            currentAgentMessageElement = null;
        }
        stopAndClearAudioQueue();

        const message = {mime_type: "text/plain", data: messageText};
        try {
            audioStreamSocket.send(JSON.stringify(message));
            addMessageToChat(messageText, 'user');
            chatMessageInput.value = '';
        } catch (e) {
            console.error("Error sending text message:", e);
            addMessageToChat("Error sending message. Please try again.", "system-error");
        }
    }

    async function handleDeleteSession() {
        if (!confirm("Are you sure you want to delete the session and clear chat history? This will disconnect the stream.")) {
            return;
        }
        teardownStreamingPipeline();
        await new Promise(resolve => setTimeout(resolve, 100));
        try {
            const response = await fetch(`${coachAgentURL}/api/v1/chat`, {method: 'DELETE'});
            if (response.ok) {
                chatHistory.innerHTML = '';
                addMessageToChat("Session cleared. You can start a new one by enabling the 'Coach Connection'.", 'system-success');
            } else {
                const text = await response.text();
                addMessageToChat(`Error deleting session: ${response.status} ${text || ''}`, 'system-error');
            }
        } catch (error) {
            addMessageToChat("Network error while deleting session.", 'system-error');
            console.error('Network error deleting session:', error);
        }
    }

    async function handleStreamToggle() {
        if (isConnecting) return;

        const isEnabled = streamToggleSwitch.checked;

        if (isEnabled) {
            if (!showTextToggleSwitch.checked && !playAudioToggleSwitch.checked) {
                addMessageToChat("Text and Audio outputs were disabled. Re-enabling both for the new session.", "system-info");
                showTextToggleSwitch.checked = true;
                playAudioToggleSwitch.checked = true;
            }

            isConnecting = true;
            streamToggleSwitch.disabled = true;
            addMessageToChat("Initializing audio connection...", "system-info");
            try {
                await setupStreamingPipeline();
                setControlsDisabled(false);

                // Automatically enable the microphone on connection.
                isMuted = false;
                micMuteToggleSwitch.checked = true;
                console.log("Connection successful. Microphone is now unmuted.");
                addMessageToChat("Audio connected. You can now speak.", "system-success");

            } catch (error) {
                console.error("Failed to establish connection:", error);
                addMessageToChat("Connection failed. Please check permissions and refresh.", "system-error");
                teardownStreamingPipeline();
            } finally {
                isConnecting = false;
                streamToggleSwitch.disabled = false;
                updateMicMuteToggleState();
            }
        } else {
            addMessageToChat("Disconnecting...", "system-info");
            teardownStreamingPipeline();
        }
    }

    function handleOutputToggleChange() {
        if (!isStreaming) return;

        if (!showTextToggleSwitch.checked && !playAudioToggleSwitch.checked) {
            addMessageToChat("Both text and audio outputs are disabled. Disconnecting.", "system-info");
            streamToggleSwitch.checked = false;
            handleStreamToggle();
        }
    }

    // --- Event Listeners ---
    streamToggleSwitch.addEventListener('change', handleStreamToggle);
    micMuteToggleSwitch.addEventListener('change', toggleMute);
    micGainSlider.addEventListener('input', () => {
        if (micGainNode) micGainNode.gain.value = parseFloat(micGainSlider.value);
    });
    chatSendButton.addEventListener('click', sendMessage);
    chatMessageInput.addEventListener('keypress', (event) => {
        if (event.key === 'Enter') sendMessage();
    });
    chatDeleteSessionButton.addEventListener('click', handleDeleteSession);
    showTextToggleSwitch.addEventListener('change', handleOutputToggleChange);
    playAudioToggleSwitch.addEventListener('change', handleOutputToggleChange);


    window.addEventListener('beforeunload', () => {
        console.log("Page is unloading. Tearing down the chat streaming pipeline.");
        teardownStreamingPipeline();
    });

    // --- Initial State ---
    setControlsDisabled(true);
    updateMicMuteToggleState();

    return {addMessageToChat};
}