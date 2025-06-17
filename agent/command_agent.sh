GOOGLE_GENAI_USE_VERTEXAI=TRUE \
GOOGLE_CLOUD_PROJECT=c4-blaquiere-sbx \
GOOGLE_CLOUD_LOCATION=europe-west1 \
COACH_BACKEND_URL=http://localhost:8080 \
adk api_server

GOOGLE_GENAI_USE_VERTEXAI=TRUE \
GOOGLE_CLOUD_PROJECT=c4-blaquiere-sbx \
GOOGLE_CLOUD_LOCATION=us-central1 \
COACH_BACKEND_URL=http://localhost:8080 \
adk api_server

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