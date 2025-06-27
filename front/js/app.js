// C:/Users/Guillou/IdeaProjects/coach/front/js/app.js

import {initPlanModule} from './plan.js';
import {initChatModule} from './chat.js';

document.addEventListener('DOMContentLoaded', () => {
    // --- Shared Configuration ---
    const coachBackendURL = "http://localhost:8080";
    const coachAgentURL = "http://localhost:8081";
    const userEmail = "guillaume.blaquiere@gmail.com";

    // --- Module Initialization ---

    // Initialize the chat module first to get the addMessageToChat function
    const {addMessageToChat} = initChatModule(coachAgentURL);

    // Initialize the plan module and pass it the shared function
    initPlanModule(coachBackendURL, userEmail, addMessageToChat);

    console.log("Application initialized.");
});