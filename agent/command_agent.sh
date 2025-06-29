GOOGLE_GENAI_USE_VERTEXAI=TRUE \
GOOGLE_CLOUD_PROJECT=<YOUR_PROJECT_ID> \
GOOGLE_CLOUD_LOCATION=us-central1 \
GEMINI_MODEL_VERSION="gemini-2.0-flash-live-preview-04-09" \
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


store_unfiltered_order_calendar
store_unfiltered_unified_order_calendar
z_prolog_order_picking_delivery_line
internal_rao_sales_forecast
fv_prolog_order_lines
logistic_replenishment_proposal
store_order_calendar
store_order_schedule
internal_rao_assortment
logistic_replenishment_proposal_historica