# Coach Application

This application is a personal coaching tool to manage and track daily stretching and exercise plans. It features a Go backend, a web frontend, and a Python-based AI agent to provide coaching feedback.

## Architecture

The application is composed of four main components that work together:

1.  **Web Frontend**: A static web application for user interaction.
2.  **Coach Backend**: A Go-based API that manages data, user plans, and real-time updates.
3.  **Agent Wrapper**: A Go-based service that acts as a proxy to the Python agent.
4.  **Agent ADK**: A Python-based agent using Vertex AI for intelligent coaching logic.

These services are designed to be run as independent microservices.

```
[ User ] <--> [ Web Frontend ] <--> [ Coach Backend (Go) ] <--> [ Google Firestore ]
                         ^                    |      ^ 
                         |                    |      | Tools (HTTP)
                         |                    |      +-------------------------+
                         |                    |                                |
                         | (WebSocket)        | (HTTP - Service-to-Service)    |
                         |                    v                                |
                         +--------> [ Agent Wrapper (Go) ]                     |
                                              | (HTTP - session)               |
                                              | (WebSocket - stream)           |
                                              v                                |
                                 [ Agent ADK (Python/Vertex AI) ] <------------+
```

## Technology Stack

-   **Backend**: Go, Chi (v5), Gorilla WebSocket
-   **Agent**: Python, FastAPI, Vertex AI SDK
-   **Frontend**: HTML, CSS, JavaScript (no framework)
-   **Database**: Google Firestore
-   **Deployment**: Docker, Google Cloud Run, Google Cloud Build, Google Artifact Registry

---

## Getting Started

### Prerequisites

-   A Google Cloud Project with billing enabled.
-   The following Google Cloud APIs enabled:
    -   Firestore API (`firestore.googleapis.com`)
    -   Vertex AI API (`aiplatform.googleapis.com`)
    -   Cloud Run API (`run.googleapis.com`)
    -   Cloud Build API (`cloudbuild.googleapis.com`)
    -   Artifact Registry API (`artifactregistry.googleapis.com`)
-   Go (version 1.24 or later)
-   Python (version 3.10 or later)
-   Docker
-   `gcloud` CLI configured to use your project.

### Configuration

Before running the application, you must replace the placeholder values in the shell scripts with your own project-specific information.

-   **Project ID**: In `coach/command.sh`, `deploy.sh`, `start_agent.sh`, and `start_back.sh`, replace `<YOUR_PROJECT_ID>` with your Google Cloud project ID.
-   **Project Number**: In `deploy.sh`, replace `<YOUR_PROJECT_NUMBER>` with your Google Cloud project number.
-   **Backend URL**: In `coach/command.sh`, replace `<YOUR_BACKEND_URL>` with the URL of your deployed backend service (after deployment).
-   **Region**: In `deploy.sh`, you may want to replace the default `us-central1` region.

### Running Locally

1.  **Authenticate gcloud**: For local development connecting to Google Cloud services, authenticate the gcloud CLI.
    ```bash
    gcloud auth application-default login
    ```

2.  **Start Services**: Open three separate terminal windows and run the start scripts. Make sure you have updated `<YOUR_PROJECT_ID>` in the scripts first.

    -   **Terminal 1: Backend**
        ```bash
        ./start_back.sh
        ```
    -   **Terminal 2: Agent ADK (Python)**
        ```bash
        ./start_agent.sh
        ```
    -   **Terminal 3: Agent Wrapper**
        ```bash
        ./start_wrapper.sh
        ```

3.  **Serve Frontend**: The frontend is in the `front/` directory and can be served with any static file server. For example, using Python:
    ```bash
    python3 -m http.server 8000 --directory front
    ```
    You can then access the application at `http://localhost:8000`.

---

## Services Deep Dive

### 1. Web Frontend

A vanilla JavaScript, HTML, and CSS single-page application that serves as the user interface.

-   **Location**: `front/`
-   **Functionality**:
    -   Displays daily training plans.
    -   Allows users to update their progress.
    -   Connects to the backend via REST API and WebSockets for real-time updates.

### 2. Coach Backend

The `coach` service is the central API for the application.

-   **Location**: `coach/`
-   **Technology**: Go, Chi router, Gorilla WebSocket, Google Firestore client.
-   **Functionality**:
    -   Provides RESTful API endpoints for managing `categories`, `drills`, and `daily-plans`.
    -   Handles user-specific data persistence in Firestore.
    -   Supports WebSocket connections for real-time client updates.
    -   Invokes the Agent Wrapper to get coaching feedback.

### 3. Agent Wrapper

The `agent/wrapper_agent` service acts as a simple proxy for the Python agent.

-   **Location**: `agent/wrapper_agent/`
-   **Technology**: Go.
-   **Functionality**: Forwards requests from the Coach Backend to the Agent ADK. This decouples the main backend from the agent's specific location and implementation.

### 4. Agent ADK

The `agent/adk` service is the "intelligent" core of the application, providing coaching advice.

-   **Location**: `agent/adk/`
-   **Technology**: Python, FastAPI.
-   **Functionality**:
    -   Exposes a FastAPI web server to receive requests.
    -   Uses the Google Vertex AI SDK to interact with a generative model.
    -   Generates coaching feedback based on user progress and predefined prompts (`prompt.py`).

---

## API and WebSockets

You can interact with the API using `curl`. The `command.sh` script provides a set of useful examples.

### API Endpoints

The backend runs on port `8080` by default.

-   `GET /api/v1/daily-plans/today`: Fetches the training plan for the current date. If no plan exists, one is created automatically.
-   `PUT /api/v1/daily-plans/today`: Updates a drill in today's plan.
-   `POST /api/v1/daily-plans/initiate`: Creates a plan for a specific date provided in the request body.
-   `GET /api/v1/categories`: Lists all exercise categories.
-   `POST /api/v1/categories`: Creates a new category.
-   `GET /api/v1/drills`: Lists all drills.
-   `POST /api/v1/drills`: Creates a new drill.

### Example: Update a Drill

To update a drill, you send a `PUT` request with a JSON payload. The `id` is a combination of the user's email and the date.

```bash
# updateDrill.json
# {
#   "id": "user@example.com-2024-07-30",
#   "date": "2024-07-30",
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

### Real-time Communication with WebSockets

The application uses WebSockets to push real-time updates to connected clients, such as the web frontend.

1.  **Connection**: A client establishes a WebSocket connection by connecting to the `/api/v1/ws?email=<user_email>` endpoint. The `email` parameter is used to identify and associate the connection with a specific user.

2.  **Update Event**: When a user's training plan is modified (e.g., via a `PUT` request to `/api/v1/daily-plans/today`), the backend triggers an update.

3.  **Message Broadcast**: The backend sends a JSON message over the WebSocket to all connected clients for that user. The message notifies the client that the plan has been updated and includes the new plan data.

The WebSocket message has the following structure:

```jsonc
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
-   `front/app.yaml`: Configuration file for deploying the frontend to Google App Engine.

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


## Next release

Here my idea to test and improve the agent:
* Add Firebase auth security
* Add Service to Service security
* Add MCP server to interact with backend
* Optimize deployment and runtime to be cost efficient with good UX
* Add responsive/mobile capabilities
* Add multi-language support
* Add dynamic template
* Miscellaneous improvements (cheer up, graphic,...)
