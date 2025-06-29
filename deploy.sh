projectNumber=<YOUR_PROJECT_NUMBER>
projectId=<YOUR_PROJECT_ID>

# Create a service account for the agent service
gcloud iam service-accounts create coach-agent-service-account \
  --display-name "Service account for Coach Agent" \
  --project=${projectId}

# Grant the service account necessary roles
gcloud projects add-iam-policy-binding ${projectId} \
  --member="serviceAccount:coach-agent-service-account@${projectId}.iam.gserviceaccount.com" \
  --role=roles/run.invoker

gcloud projects add-iam-policy-binding ${projectId} \
  --member="serviceAccount:coach-agent-service-account@${projectId}.iam.gserviceaccount.com" \
  --role=roles/aiplatform.user

#gcloud projects add-iam-policy-binding ${projectId} \
#  --member="serviceAccount:coach-agent-service-account@${projectId}.iam.gserviceaccount.com" \
#  --role=roles/texttospeech.user

# Create a service account for the backend service
gcloud iam service-accounts create coach-backend-service-account \
  --display-name "Service account for Coach Backend" \
  --project=${projectId}

# Grant the service account necessary roles
gcloud projects add-iam-policy-binding ${projectId} \
  --member="serviceAccount:coach-backend-service-account@${projectId}.iam.gserviceaccount.com" \
  --role=roles/datastore.user


# Build the containers
gcloud builds submit --project=${projectId} --tag gcr.io/${projectId}/coach-backend:latest ./coach
gcloud builds submit --project=${projectId} --tag gcr.io/${projectId}/coach-agent:latest ./agent/adk
gcloud builds submit --project=${projectId} --tag gcr.io/${projectId}/coach-agent-wrapper:latest ./agent/wrapper_agent
gcloud builds submit --project=${projectId} --tag gcr.io/${projectId}/coach-frontend:latest ./front
# Deploy the backend service
gcloud run deploy coach-backend \
  --image gcr.io/${projectId}/coach-backend:latest \
  --platform managed \
  --region europe-west1 \
  --allow-unauthenticated \
  --port 8080 \
  --memory 512Mi \
  --cpu 1 \
  --set-env-vars=PROJECT_ID=${projectId} \
  --service-account=coach-backend-service-account@${projectId}.iam.gserviceaccount.com


# Deploy the agent service as a sidecar of the wrapper service
gcloud beta run deploy coach-agent-wrapper \
  --platform managed \
  --region europe-west1 \
  --allow-unauthenticated \
  --service-account=coach-agent-service-account@${projectId}.iam.gserviceaccount.com \
  --container wrapper \
  --memory 512Mi \
  --cpu 1 \
  --image gcr.io/${projectId}/coach-agent-wrapper:latest \
  --port 8081 \
  --set-env-vars=COACH_AGENT_HOST=localhost,COACH_AGENT_PORT=8000,COACH_AGENT_NAME=coach_agent,COACH_BACKEND_URL=http://<YOUR_BACKEND_URL> \
  --depends-on=adk \
  --container adk \
  --memory 512Mi \
  --cpu 2 \
  --image gcr.io/${projectId}/coach-agent:latest \
  --startup-probe=httpGet.path=/list-apps,httpGet.port=8000,timeoutSeconds=10,periodSeconds=10,failureThreshold=3,initialDelaySeconds=5 \
  --set-env-vars=GOOGLE_GENAI_USE_VERTEXAI=TRUE,GOOGLE_CLOUD_PROJECT=${projectId},GOOGLE_CLOUD_LOCATION=europe-west1,COACH_BACKEND_URL=http://<YOUR_BACKEND_URL>

# Deploy the frontend service
gcloud run deploy coach-frontend \
  --image gcr.io/${projectId}/coach-frontend:latest \
  --platform managed \
  --region europe-west1 \
  --allow-unauthenticated \
  --port 80 \
  --memory 256Mi \
  --cpu 1


# Create endpoint NEG for the 3 services
gcloud compute network-endpoint-groups create coach-backend-neg \
  --project=${projectId} \
  --region=europe-west1 \
  --network-endpoint-type=serverless \
  --cloud-run-service=coach-backend

gcloud compute network-endpoint-groups create coach-agent-wrapper-neg \
  --project=${projectId} \
  --region=europe-west1 \
  --network-endpoint-type=serverless \
  --cloud-run-service=coach-agent-wrapper

gcloud compute network-endpoint-groups create coach-frontend-neg \
  --project=${projectId} \
  --region=europe-west1 \
  --network-endpoint-type=serverless \
  --cloud-run-service=coach-frontend

# Create the backend service for the 3 services
gcloud compute backend-services create coach-backend-service \
  --project=${projectId} \
  --protocol=HTTP \
  --global

gcloud compute backend-services create coach-agent-wrapper-service \
  --project=${projectId} \
  --protocol=HTTP \
  --global

gcloud compute backend-services create coach-frontend-service \
  --project=${projectId} \
  --protocol=HTTP \
  --global

# Add the NEG to the backend service
gcloud compute backend-services add-backend coach-backend-service \
  --project=${projectId} \
  --global \
  --network-endpoint-group=coach-backend-neg \
  --network-endpoint-group-region=europe-west1

gcloud compute backend-services add-backend coach-agent-wrapper-service \
  --project=${projectId} \
  --global \
  --network-endpoint-group=coach-agent-wrapper-neg \
  --network-endpoint-group-region=europe-west1

gcloud compute backend-services add-backend coach-frontend-service \
  --project=${projectId} \
  --global \
  --network-endpoint-group=coach-frontend-neg \
  --network-endpoint-group-region=europe-west1

# Create a URL map to route requests to the correct backend service
gcloud compute url-maps create coach-url-map \
  --project=${projectId} \
  --default-service coach-frontend-service

# Add path matchers for the backend and agent services
gcloud compute url-maps add-path-matcher coach-url-map \
  --project=${projectId} \
  --default-service coach-frontend-service \
  --path-matcher-name=path-matcher-coach \
  --path-rules="/api/v1/daily-plans/*=coach-backend-service,/api/v1/categories/*=coach-backend-service,/api/v1/drills/*=coach-backend-service,/api/v1/plan-templates/*=coach-backend-service,/health=coach-backend-service,/api/v1/chat=coach-agent-wrapper-service"

# Create a target HTTP proxy
gcloud compute target-http-proxies create coach-http-proxy \
  --project=${projectId} \
  --url-map coach-url-map

# Create a global forwarding rule
gcloud compute forwarding-rules create coach-http-rule \
  --project=${projectId} \
  --global \
  --ports 80 \
  --target-http-proxy coach-http-proxy

# Get the IP address of the forwarding rule
gcloud compute forwarding-rules list --project=${projectId} --global
