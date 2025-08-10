GOOGLE_GENAI_USE_VERTEXAI=TRUE \
GOOGLE_CLOUD_PROJECT=<YOUR_PROJECT_ID> \
GOOGLE_CLOUD_LOCATION=<YOUR_REGION> \
GOOGLE_AGENT_ENGINE_ID=<AGENT_ENGINE_ID> \
GOOGLE_AGENT_ENGINE_LOCATION=<AGENT_ENGINE_LOCATION> \
GEMINI_MODEL_VERSION="gemini-live-2.5-flash-preview-native-audio" \
COACH_BACKEND_URL=http://localhost:8080 \
uvicorn main:app --host 0.0.0.0 --port 8000 --reload

curl http://localhost:8000/apps/coach_agent/users/u_123/sessions/s_123

curl -X POST http://localhost:8000/apps/coach_agent/users/u_123/sessions/s_123 \
  -H "Content-Type: application/json"

curl -X POST http://localhost:8000/run \
-H "Content-Type: application/json" \
-d '{
"appName": "coach_agent",
"userId": "u_123",
"sessionId": "s_123",
"newMessage": {
    "role": "user",
    "parts": [{
    "text": "Bonjour, on fait quoi aujourd hui ?"
    }]
}
}'



curl -X POST http://localhost:8000/run \
-H "Content-Type: application/json" \
-d "{
\"appName\": \"coach_agent\",
\"userId\": \"u_123\",
\"sessionId\": \"s_123\",
\"newMessage\": {
    \"role\": \"user\",
    \"parts\": [{
    \"inlineData\": {
     \"data\": \"$(cat hello_base64.txt)\",
     \"mimeType\": \"audio/m4a\"
     }
    }]
}
}"

curl -X POST http://localhost:8000/run \
-H "Content-Type: application/json" \
-d "{
\"appName\": \"coach_agent\",
\"userId\": \"u_123\",
\"sessionId\": \"s_123\",
\"newMessage\": {
    \"role\": \"user\",
    \"parts\": [{
    \"inlineData\": {
     \"data\": \"\",
     \"mimeType\": \"audio/m4a\"
     }
    }]
}
}"