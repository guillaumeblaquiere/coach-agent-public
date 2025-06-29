# Coach Application

This application is a personal coaching tool designed to manage and track daily stretching and exercise plans. It provides a backend service to manage the data, a web frontend for user interaction, and agent components to offer coaching functionalities.

## Architecture

The application is composed of three main services:

1.  **Coach Backend**: A Go-based backend that serves the main API for managing exercise categories, drills, and daily training plans. It uses Google Firestore for data persistence and WebSockets for real-time updates.
2.  **Agent Wrapper**: A Go-based service that acts as a wrapper or intermediary for the main agent, likely handling orchestration between the backend and the agent's logic.
3.  **Agent ADK**: A Python-based agent that contains the core coaching logic, processing user input and progress to provide feedback.

These services are designed to be run independently and communicate with each other, likely via API calls.

---

## Getting Started

### Prerequisites

-   Go (version 1.24 or later)
-   Python (version 3.10 or later)
-   Docker
-   A Google Cloud Project with the following APIs enabled:
    -   Firestore API
    -   Vertex AI API
-   `gcloud` CLI configured to use your project.

### Configuration

Before running the application, you must replace the placeholder values in the shell scripts with your own project-specific information.

-   **Project ID**: In `coach/command.sh`, `deploy.sh`, `start_agent.sh`, and `start_back.sh`, replace `<YOUR_PROJECT_ID>` with your Google Cloud project ID.
-   **Project Number**: In `deploy.sh`, replace `<YOUR_PROJECT_NUMBER>` with your Google Cloud project number.
-   **Backend URL**: In `coach/command.sh` and `deploy.sh`, replace `<YOUR_BACKEND_URL>` with the URL of your deployed backend service.

### Running Locally

Each service can be started using the provided shell scripts:

-   **Backend**: `./start_back.sh`
-   **Agent Wrapper**: `./start_wrapper.sh`
-   **Agent ADK**: `./start_agent.sh`

The frontend can be served from the `front/` directory using any static file server.

---

## Services

### 1. Coach Backend

The `coach` service is the central API for the application.

-   **Technology**: Go, Chi router, Gorilla WebSocket, Google Firestore client.
-   **Functionality**:
    -   Provides RESTful API endpoints for managing `categories`, `drills`, and `daily-plans`.
    -   Handles user-specific data persistence in Firestore.
    -   Supports WebSocket connections for real-time client updates.

### 2. Agent Wrapper

The `agent/wrapper_agent` service acts as a controller or proxy for the Python agent.

-   **Technology**: Go.
-   **Functionality**: Manages and communicates with the `Agent ADK`, orchestrating the flow of information between the main backend and the agent's logic.

### 3. Agent ADK

The `agent/adk` service is the "intelligent" core of the application.

-   **Technology**: Python.
-   **Functionality**: Contains the business logic for coaching. The `coach_agent` subdirectory (`agent.py`, `prompt.py`) suggests it processes prompts and executes agent tasks based on user interaction.

---

## API Usage and Interaction Examples

You can interact with the API using `curl`. The `coach/command.sh` script provides a good starting point.

**Example 1: Get today's training plan**

This command fetches the training plan for the current date. If no plan exists, the API automatically creates one based on the default template.

```bash
curl http://localhost:8080/api/v1/daily-plans/today
```

**Example 2: Initiate a plan for a specific date**

This is useful for pre-populating plans for future dates.

```bash
curl -X POST http://localhost:8080/api/v1/daily-plans/initiate
```

**Example 3: Update a drill in today's plan**

To update a drill, you send a `PUT` request with a JSON payload. This example updates the "Abs" drill. The file `updateDrill.json` contains the data to be updated.

```bash
# updateDrill.json
# {
#   "id": "<YOUR_EMAIL>-2025-06-01",
#   "date": "2025-06-01",
#   "repetitions": {
#     "back": {
#       "Abs": {
#         "repetition": 3
#       }
#     }
#   }
# }

curl -X PUT -H "Content-Type: application/json" -d "@updateDrill.json" http://localhost:8080/api/v1/daily-plans/today
```

---

## Real-time Communication with WebSockets

The application uses WebSockets to push real-time updates to connected clients, such as the web frontend.

### How it Works

1.  **Connection**: A client establishes a WebSocket connection by connecting to the `/api/v1/ws?email=<user_email>` endpoint. The `email` parameter is used to identify and associate the connection with a specific user.

2.  **Update Event**: When a user's training plan is modified (e.g., via a `PUT` request to `/api/v1/daily-plans/today`), the backend triggers an update.

3.  **Message Broadcast**: The backend sends a JSON message over the WebSocket to all connected clients for that user. The message notifies the client that the plan has been updated and includes the new plan data.

### Message Format

The WebSocket message has the following structure:

```json
{
  "action": "PLAN_UPDATED",
  "data": { ... updated DailyTrainingPlan object ... },
  "source": "api"
}
```

-   `action`: Describes the type of event (e.g., `PLAN_UPDATED`).
-   `data`: Contains the payload, which is the full, updated training plan.
-   `source`: Indicates what triggered the update (e.g., `api`, `agent`).

This mechanism allows the UI to refresh instantly without needing to poll the server for changes.

---

## Deployment

The application is designed to be deployed using Docker. Each service has its own `Dockerfile` for building a container image.

### Docker Images

-   `coach/Dockerfile`: A multi-stage Dockerfile that builds the Go backend and creates a minimal final image from `debian:buster-slim`.
-   `agent/adk/Dockerfile`: Builds the Python agent service.
-   `agent/wrapper_agent/Dockerfile`: Builds the Go-based agent wrapper.
-   `front/Dockerfile`: (Assumed) Builds a static web server (e.g., Nginx) to serve the frontend files.

### Deployment Process

The `deploy.sh` script is provided to automate the deployment. Before running it, ensure you have updated the placeholder values as described in the **Configuration** section.

A typical process would be:

1.  **Build Images**: Build the Docker image for each service.
    ```bash
    docker build -t gcr.io/<YOUR_PROJECT_ID>/coach-backend -f coach/Dockerfile .
    docker build -t gcr.io/<YOUR_PROJECT_ID>/agent-adk -f agent/adk/Dockerfile .
    # ... and so on for other services
    ```
2.  **Push Images**: Push the images to a container registry like Google Container Registry (GCR).
    ```bash
    docker push gcr.io/<YOUR_PROJECT_ID>/coach-backend
    docker push gcr.io/<YOUR_PROJECT_ID>/agent-adk
    ```
3.  **Deploy Services**: Deploy the images to a cloud platform like Google Cloud Run, which is suggested by the use of `PROJECT_ID` and `PORT` environment variables.

Please inspect the `deploy.sh` script for the specific commands and configurations used in this project.
