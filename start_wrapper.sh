cd agent/wrapper_agent
COACH_AGENT_PORT=8000 \
COACH_AGENT_HOST=localhost \
COACH_AGENT_NAME="coach_agent" \
COACH_BACKEND_URL=http://localhost:8080 \
go run .