COACH_AGENT_LOCAL_PORT=8000 \
COACH_AGENT_NAME="coach_agent" \
go run .


# Curl command with audio file
curl -X POST http://localhost:8081/api/v1/chat \
-H "Content-Type: application/json" \
-d "{
\"inlineData\": {
 \"data\": \"$(cat hello_base64.txt)\",
 \"mimeType\": \"audio/m4a\"
 }
}"

# Curl command with text message
curl -X POST http://localhost:8081/api/v1/chat \
-H "Content-Type: application/json" \
-d "{
\"text\": \"Bonjour, on fait quoi aujourd hui ?\"
}"

    <div id="chat-input-area">
        <!-- NOUVEAUX ÉLÉMENTS POUR LE SLIDER VAD -->
        <div id="vad-controls" style="display: flex; flex-direction: column; margin-bottom: 5px; width: 100%;">
            <label for="vad-threshold-slider" style="font-size: 0.8em; margin-bottom: 2px;">Seuil VAD / Gain:</label>
            <div style="display: flex; align-items: center;">
                <input type="range" id="vad-threshold-slider" min="0.1" max="5" step="0.1" value="0.8" style="flex-grow: 1; margin-right: 5px;">
                <span id="vad-threshold-value" style="font-size: 0.8em; min-width: 25px; text-align: right;">0.8</span>
            </div>
            <progress id="mic-gain-meter" value="0" max="255" style="width: 100%; height: 5px; margin-top: 2px;"></progress>
        </div>
        <!-- FIN DES NOUVEAUX ÉLÉMENTS -->
        <input type="text" id="chat-message-input" placeholder="Type a message...">
        <button id="chat-send-button" title="Send">➔</button>
        <button id="chat-mic-button" title="Toggle Microphone">🎤</button>
    </div>

