// This function will be exported and called by app.js
export function initPlanModule(coachBackendURL, userEmail, addMessageToChat) {
    const dateDisplay = document.getElementById('date-display');
    const datePicker = document.getElementById('date-picker');
    const prevDayButton = document.getElementById('prev-day');
    const nextDayButton = document.getElementById('next-day');
    const planContainer = document.getElementById('plan-container');
    const source = "?source=web";

    let currentDate = new Date();
    let isToday = true;
    let currentPlanTemplate = null;
    let currentDailyPlanData = null;

    // --- WebSocket State ---
    let planWebSocket = null;
    let reconnectTimer = null;

    function formatDate(date) {
        return date.toISOString().slice(0, 10);
    }

    function updateDateDisplay() {
        dateDisplay.textContent = formatDate(currentDate);
        isToday = formatDate(currentDate) === formatDate(new Date());
        planContainer.classList.toggle('read-only', !isToday);
    }

    async function loadPlanTemplate() {
        if (currentPlanTemplate) return Promise.resolve(currentPlanTemplate);
        try {
            const response = await fetch(`${coachBackendURL}/api/v1/plan-templates/default`);
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(`Template fetch failed: ${response.statusText} (${errorData.error || 'Unknown error'})`);
            }
            currentPlanTemplate = await response.json();
            return currentPlanTemplate;
        } catch (error) {
            console.error('Error loading plan template:', error);
            planContainer.innerHTML = `<p>Error loading plan structure: ${error.message}</p>`;
            throw error;
        }
    }

    async function loadDailyPlanData(dateStr) {
        const endpoint = `/api/v1/daily-plans/${dateStr === formatDate(new Date()) ? 'today' : dateStr}`;
        try {
            let response = await fetch(coachBackendURL + endpoint);
            if (response.status === 404 && dateStr === formatDate(new Date())) {
                const initResponse = await fetch(`${coachBackendURL}/api/v1/daily-plans/initiate`, {method: 'POST'});
                if (!initResponse.ok) {
                    const errorData = await initResponse.json().catch(() => ({}));
                    throw new Error(`Failed to initiate plan: ${initResponse.statusText} (${errorData.error || 'Unknown error'})`);
                }
                currentDailyPlanData = await initResponse.json();
            } else if (response.status === 404) {
                currentDailyPlanData = null;
            } else if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(`Failed to load daily plan: ${response.statusText} (${errorData.error || 'Plan not found'})`);
            } else {
                currentDailyPlanData = await response.json();
            }
            return currentDailyPlanData;
        } catch (error) {
            console.error('Error loading daily plan data:', error);
            planContainer.innerHTML = `<p>Error loading plan data: ${error.message}</p>`;
            currentDailyPlanData = null;
            throw error;
        }
    }

    async function loadPlan() {
        const dateStr = formatDate(currentDate);
        planContainer.innerHTML = '<p>Loading plan...</p>';
        updateDateDisplay();

        try {
            await loadPlanTemplate();
            await loadDailyPlanData(dateStr);
            if (currentPlanTemplate) {
                displayPlan(currentPlanTemplate, currentDailyPlanData);
            }
        } catch (error) {
            console.error('Error in loadPlan orchestration:', error);
            if (!planContainer.innerHTML.includes("Error") && !planContainer.innerHTML.includes("Loading")) {
                planContainer.innerHTML = `<p>An unexpected error occurred.</p>`;
            }
        }
    }

    function displayPlan(planTemplate, dailyPlanData) {
        planContainer.innerHTML = '';
        if (!planTemplate || !planTemplate.categories) {
            planContainer.innerHTML = '<p>No plan structure (template) available to display.</p>';
            return;
        }

        for (const categoryId in planTemplate.categories) {
            const category = planTemplate.categories[categoryId];
            const categoryElement = document.createElement('div');
            categoryElement.classList.add('accordion');
            categoryElement.dataset.categoryId = categoryId;

            const header = document.createElement('div');
            header.classList.add('accordion-header');
            header.innerHTML = `<h3>${category.name}</h3>`;

            const categoryRepsDisplay = document.createElement('span');
            categoryRepsDisplay.classList.add('category-reps');
            header.appendChild(categoryRepsDisplay);

            const catButtons = document.createElement('div');
            const catIncButton = document.createElement('button');
            catIncButton.textContent = '+';
            catIncButton.onclick = (e) => {
                e.stopPropagation();
                adjustCategoryReps(categoryId, 1);
            };
            const catDecButton = document.createElement('button');
            catDecButton.textContent = '-';
            catDecButton.onclick = (e) => {
                e.stopPropagation();
                adjustCategoryReps(categoryId, -1);
            };
            catButtons.append(catIncButton, catDecButton);
            header.appendChild(catButtons);
            categoryElement.appendChild(header);

            const content = document.createElement('div');
            content.classList.add('accordion-content', 'show');

            let minRepsInCategory = Infinity;
            let hasDrillsInCat = false;

            if (category.drills) {
                for (const drillId in category.drills) {
                    hasDrillsInCat = true;
                    const drill = category.drills[drillId];
                    const drillElement = document.createElement('div');
                    drillElement.classList.add('drill');
                    drillElement.dataset.drillId = drillId;
                    drillElement.dataset.categoryId = categoryId;

                    const drillName = document.createElement('div');
                    drillName.classList.add('drill-name');
                    drillName.textContent = drill.name;
                    drillElement.appendChild(drillName);

                    if (drill.description) {
                        const drillDesc = document.createElement('div');
                        drillDesc.classList.add('drill-desc');
                        drillDesc.textContent = drill.description;
                        drillElement.appendChild(drillDesc);
                    }

                    const drillControls = document.createElement('div');
                    drillControls.classList.add('drill-controls');
                    const decButton = document.createElement('button');
                    decButton.textContent = '-';
                    decButton.onclick = () => adjustDrillReps(categoryId, drillId, -1);
                    const incButton = document.createElement('button');
                    incButton.textContent = '+';
                    incButton.onclick = () => adjustDrillReps(categoryId, drillId, 1);
                    const repsDisplay = document.createElement('span');
                    repsDisplay.classList.add('drill-repetition');

                    let currentRepetition = 0;
                    let currentNote = "";
                    if (dailyPlanData && dailyPlanData.repetitions && dailyPlanData.repetitions[categoryId] && dailyPlanData.repetitions[categoryId][drillId]) {
                        const achievement = dailyPlanData.repetitions[categoryId][drillId];
                        currentRepetition = achievement.repetition || 0;
                        currentNote = achievement.note || "";
                    }
                    repsDisplay.textContent = currentRepetition;
                    minRepsInCategory = Math.min(minRepsInCategory, currentRepetition);

                    const noteInput = document.createElement('input');
                    noteInput.type = 'text';
                    noteInput.classList.add('drill-note-input');
                    noteInput.placeholder = 'Notes...';
                    noteInput.value = currentNote;

                    drillControls.append(decButton, repsDisplay, incButton, noteInput);
                    drillElement.appendChild(drillControls);
                    content.appendChild(drillElement);
                }
            }
            categoryRepsDisplay.textContent = `(Min Reps: ${hasDrillsInCat && minRepsInCategory !== Infinity ? minRepsInCategory : 0})`;
            categoryElement.appendChild(content);
            planContainer.appendChild(categoryElement);

            header.addEventListener('click', (e) => {
                if (e.target.closest('button')) return;
                content.classList.toggle('show');
            });
        }
        updateDateDisplay();
    }

    function ensureDailyPlanDataInitialized() {
        if (!currentDailyPlanData) {
            if (isToday) {
                currentDailyPlanData = {
                    id: `${formatDate(currentDate)}-temp`,
                    templateId: currentPlanTemplate?.id || "default",
                    date: formatDate(currentDate),
                    repetitions: {},
                    createdAt: new Date().toISOString(),
                    updatedAt: new Date().toISOString()
                };
            } else {
                console.warn("Cannot modify: No existing daily plan data for this past date.");
                addMessageToChat("Cannot modify plan for a past date without existing data.", "system-warning");
                return false;
            }
        }
        if (!currentDailyPlanData.repetitions) {
            currentDailyPlanData.repetitions = {};
        }
        return true;
    }

    function adjustDrillReps(categoryId, drillId, adjustment) {
        if (!currentPlanTemplate || !currentPlanTemplate.categories[categoryId]?.drills[drillId]) {
            console.warn("Drill definition not found in template.");
            return;
        }
        if (!isToday) {
            addMessageToChat("Modifications are only allowed for today's plan.", "system-info");
            return;
        }
        if (!ensureDailyPlanDataInitialized()) return;

        if (!currentDailyPlanData.repetitions[categoryId]) currentDailyPlanData.repetitions[categoryId] = {};
        if (!currentDailyPlanData.repetitions[categoryId][drillId]) currentDailyPlanData.repetitions[categoryId][drillId] = {
            repetition: 0,
            note: ""
        };

        let achievement = currentDailyPlanData.repetitions[categoryId][drillId];
        let newReps = (achievement.repetition || 0) + adjustment;
        if (newReps < 0) newReps = 0;

        const drillElement = planContainer.querySelector(`.drill[data-category-id="${categoryId}"][data-drill-id="${drillId}"]`);
        let noteFromInput = achievement.note;
        if (drillElement) {
            const noteInputElement = drillElement.querySelector('.drill-note-input');
            if (noteInputElement) noteFromInput = noteInputElement.value;
        }

        achievement.repetition = newReps;
        achievement.note = noteFromInput;

        displayPlan(currentPlanTemplate, currentDailyPlanData);
        saveDrillUpdate(categoryId, drillId, newReps, achievement.note);
    }

    function adjustCategoryReps(categoryId, adjustment) {
        if (!currentPlanTemplate || !currentPlanTemplate.categories[categoryId]?.drills) {
            console.warn("Category or drills not found in template.");
            return;
        }
        if (!isToday) {
            addMessageToChat("Modifications are only allowed for today's plan.", "system-info");
            return;
        }
        if (!ensureDailyPlanDataInitialized()) return;

        const categoryDrills = currentPlanTemplate.categories[categoryId].drills;
        const drillsToUpdate = {};
        if (!currentDailyPlanData.repetitions[categoryId]) currentDailyPlanData.repetitions[categoryId] = {};
        if (!drillsToUpdate[categoryId]) drillsToUpdate[categoryId] = {};

        for (const drillId in categoryDrills) {
            if (!currentDailyPlanData.repetitions[categoryId][drillId]) currentDailyPlanData.repetitions[categoryId][drillId] = {
                repetition: 0,
                note: ""
            };
            let achievement = currentDailyPlanData.repetitions[categoryId][drillId];
            let newReps = (achievement.repetition || 0) + adjustment;
            if (newReps < 0) newReps = 0;

            const drillElement = planContainer.querySelector(`.drill[data-category-id="${categoryId}"][data-drill-id="${drillId}"]`);
            let noteFromInput = achievement.note;
            if (drillElement) {
                const noteInputElement = drillElement.querySelector('.drill-note-input');
                if (noteInputElement) noteFromInput = noteInputElement.value;
            }

            achievement.repetition = newReps;
            achievement.note = noteFromInput;
            drillsToUpdate[categoryId][drillId] = {repetition: newReps, note: achievement.note};
        }

        displayPlan(currentPlanTemplate, currentDailyPlanData);

        if (Object.keys(drillsToUpdate[categoryId]).length > 0) {
            const payloadForBackend = {[categoryId]: drillsToUpdate[categoryId]};
            saveMultipleDrillUpdates(payloadForBackend);
        }
    }

    async function saveDrillUpdate(categoryId, drillId, newReps, note) {
        if (!isToday) {
            addMessageToChat("Modifications are only allowed for today's plan.", "system-info");
            return;
        }
        if (!ensureDailyPlanDataInitialized()) return;

        const payload = {
            repetitions: {
                [categoryId]: {
                    [drillId]: {
                        repetition: newReps,
                        note: note || ""
                    }
                }
            }
        };

        try {
            const response = await fetch(`${coachBackendURL}/api/v1/daily-plans/today${source}`, {
                method: 'PUT',
                headers: {'Content-Type': 'application/json', 'X-User-Email': userEmail},
                body: JSON.stringify(payload)
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(`Failed to save drill update: ${response.statusText} (${errorData.error || 'Unknown error'})`);
            }
            currentDailyPlanData = await response.json();
            displayPlan(currentPlanTemplate, currentDailyPlanData);
            addMessageToChat("Plan updated successfully.", "system-success");
        } catch (error) {
            console.error('Error saving drill update:', error);
            addMessageToChat(`Error saving update: ${error.message}`, 'system-error');
        }
    }

    async function saveMultipleDrillUpdates(repetitionsPayload) {
        if (!isToday || Object.keys(repetitionsPayload).length === 0) {
            if (Object.keys(repetitionsPayload).length === 0) console.log("No updates to send for category adjustment.");
            return;
        }
        if (!ensureDailyPlanDataInitialized()) return;

        const payload = {repetitions: repetitionsPayload};

        try {
            const response = await fetch(`${coachBackendURL}/api/v1/daily-plans/today`, {
                method: 'PUT',
                headers: {'Content-Type': 'application/json', 'X-User-Email': userEmail},
                body: JSON.stringify(payload)
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(`Failed to save multiple drill updates: ${response.statusText} (${errorData.error || 'Unknown error'})`);
            }
            currentDailyPlanData = await response.json();
            displayPlan(currentPlanTemplate, currentDailyPlanData);
            addMessageToChat("Category repetitions updated successfully.", "system-success");
        } catch (error) {
            console.error('Error saving multiple drill updates:', error);
            addMessageToChat(`Error saving category update: ${error.message}`, 'system-error');
        }
    }

    function connectRealtimePlanUpdates() {
        // Prevent duplicate connections
        if (planWebSocket && (planWebSocket.readyState === WebSocket.OPEN || planWebSocket.readyState === WebSocket.CONNECTING)) {
            console.log("Real-time plan update connection is already open or connecting.");
            return;
        }

        const wsURL = `${coachBackendURL.replace('http', 'ws')}/api/v1/ws?email=${encodeURIComponent(userEmail)}`;
        console.log("Connecting to WebSocket for plan updates:", wsURL);
        planWebSocket = new WebSocket(wsURL);

        planWebSocket.onopen = () => {
            console.log("Real-time plan update connection established.");
            addMessageToChat("Real-time connection active.", "system-success");
        };

        planWebSocket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                if (message.action === 'PLAN_UPDATED') {
                    addMessageToChat("Your plan has been updated in real-time!", "system-info");
                    currentDailyPlanData = message.data;
                    if (currentPlanTemplate) {
                        displayPlan(currentPlanTemplate, currentDailyPlanData);
                    }
                }
            } catch (error) {
                console.error("Error parsing WebSocket message:", error);
            }
        };

        planWebSocket.onclose = (event) => {
            console.log("Real-time plan update connection closed.", event);
            // Do not reconnect if the socket was closed intentionally by teardownPlanWebSocket
            if (planWebSocket) {
                addMessageToChat("Real-time connection lost. Attempting to reconnect...", "system-error");
                clearTimeout(reconnectTimer);
                reconnectTimer = setTimeout(connectRealtimePlanUpdates, 5000);
            } else {
                console.log("Intentional close. Not reconnecting plan WebSocket.");
            }
        };

        planWebSocket.onerror = (error) => {
            console.error("WebSocket error:", error);
            addMessageToChat("A real-time connection error occurred.", "system-error");
            // The onclose handler will be triggered after an error, managing the reconnection logic.
        };
    }

    function teardownPlanWebSocket() {
        console.log("Tearing down plan WebSocket due to page unload.");
        if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
        }
        if (planWebSocket) {
            const socketToClose = planWebSocket;
            // Signal to onclose that this is an intentional closure by setting the main variable to null
            planWebSocket = null;
            if (socketToClose.readyState === WebSocket.OPEN || socketToClose.readyState === WebSocket.CONNECTING) {
                socketToClose.close(1000, "Page unloading");
            }
        }
    }

    // --- Event Listeners ---
    dateDisplay.addEventListener('click', () => {
        datePicker.style.display = 'block';
        datePicker.value = formatDate(currentDate);
        datePicker.focus();
    });
    datePicker.addEventListener('blur', () => {
        setTimeout(() => {
            if (document.activeElement !== datePicker) datePicker.style.display = 'none';
        }, 100);
    });
    datePicker.addEventListener('change', () => {
        currentDate = new Date(datePicker.value + 'T00:00:00');
        datePicker.style.display = 'none';
        loadPlan();
    });
    prevDayButton.addEventListener('click', () => {
        currentDate.setDate(currentDate.getDate() - 1);
        loadPlan();
    });
    nextDayButton.addEventListener('click', () => {
        currentDate.setDate(currentDate.getDate() + 1);
        loadPlan();
    });

    // Add listener to clean up WebSocket on page unload
    window.addEventListener('beforeunload', teardownPlanWebSocket);

    // --- Initial Load ---
    loadPlan();
    connectRealtimePlanUpdates();
}