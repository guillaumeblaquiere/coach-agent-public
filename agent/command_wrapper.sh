COACH_AGENT_PORT=8000 \
COACH_AGENT_HOST=localhost \
COACH_BACKEND_URL=http://localhost:8080 \
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


