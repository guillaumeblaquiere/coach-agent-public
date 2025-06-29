cd agent/adk
GOOGLE_GENAI_USE_VERTEXAI=TRUE \
GOOGLE_CLOUD_PROJECT=c4-blaquiere-sbx \
GOOGLE_CLOUD_LOCATION=us-central1 \
GEMINI_MODEL_VERSION="gemini-live-2.5-flash-preview-native-audio" \
COACH_BACKEND_URL=http://localhost:8080 \
uvicorn main:app --host 0.0.0.0 --port 8000 --reload
#GEMINI_MODEL_VERSION="gemini-2.0-flash-live-preview-04-09" \
