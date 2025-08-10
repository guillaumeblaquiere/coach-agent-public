#!/bin/bash

projectNumber=<YOUR_PROJECT_NUMBER>
projectId=<YOUR_PROJECT_ID>
domainName="<YOUR_DOMAIN_NAME>"
region="<YOUR_REGION>"
agent_name="coach_agent" #"<YOUR_AGENT_NAME>"
agent_engine_location="<YOUR_AGENT_ENGINE_LOCATION>"

echo "Creating App Engine application in region ${region} if it doesn't exist..."
gcloud app create --region=${region} --project=${projectId}


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

# Create a service account for the backend service
gcloud iam service-accounts create coach-backend-service-account \
  --display-name "Service account for Coach Backend" \
  --project=${projectId}

# Grant the service account necessary roles
gcloud projects add-iam-policy-binding ${projectId} \
  --member="serviceAccount:coach-backend-service-account@${projectId}.iam.gserviceaccount.com" \
  --role=roles/datastore.user

# Create the agent engine instance
curl -H "Content-Type: application/json" -H "Authorization: Bearer $(gcloud auth print-access-token)" -X POST -d "
{
  \"name\": \"${agent_name}\",
  \"displayName\": \"${agent_name}\",
  \"description\": \"${agent_name}\",
}" \
https://${agent_engine_location}-aiplatform.googleapis.com/v1/projects/${projectId}/locations/${agent_engine_location}/reasoningEngines

agent_engine_id=$(curl -H "Content-Type: application/json" -H "Authorization: Bearer $(gcloud auth print-access-token)" \
 https://${agent_engine_location}-aiplatform.googleapis.com/v1/projects/${projectId}/locations/${agent_engine_location}/reasoningEngines | \
 jq ".reasoningEngines[] | select(.displayName == \"${agent_name}\")" | jq -r .name | awk -F'/' '{print $NF}')

# Build the containers for backend and agent services
gcloud builds submit --project=${projectId} --tag gcr.io/${projectId}/coach-backend:latest ./coach
gcloud builds submit --project=${projectId} --tag gcr.io/${projectId}/coach-agent:latest ./agent/adk
gcloud builds submit --project=${projectId} --tag gcr.io/${projectId}/coach-agent-wrapper:latest ./agent/wrapper_agent

# Deploy the backend service to Cloud Run
gcloud run deploy coach-backend \
  --image gcr.io/${projectId}/coach-backend:latest \
  --platform managed \
  --region ${region} \
  --allow-unauthenticated \
  --port 8080 \
  --memory 512Mi \
  --cpu 1 \
  --set-env-vars=PROJECT_ID=${projectId} \
  --service-account=coach-backend-service-account@${projectId}.iam.gserviceaccount.com


# Deploy the agent service as a sidecar of the wrapper service to Cloud Run
gcloud beta run deploy coach-agent-wrapper \
  --platform managed \
  --region ${region} \
  --allow-unauthenticated \
  --service-account=coach-agent-service-account@${projectId}.iam.gserviceaccount.com \
  --container wrapper \
  --memory 512Mi \
  --cpu 1 \
  --image gcr.io/${projectId}/coach-agent-wrapper:latest \
  --port 8081 \
  --set-env-vars=COACH_AGENT_HOST=localhost,COACH_AGENT_PORT=8000,COACH_AGENT_NAME=${agent_name},COACH_BACKEND_URL=http://<YOUR_BACKEND_URL> \
  --depends-on=adk \
  --container adk \
  --memory 512Mi \
  --cpu 2 \
  --image gcr.io/${projectId}/coach-agent:latest \
  --startup-probe=httpGet.path=/list-apps,httpGet.port=8000,timeoutSeconds=10,periodSeconds=10,failureThreshold=3,initialDelaySeconds=5 \
  --set-env-vars=GOOGLE_GENAI_USE_VERTEXAI=TRUE,GOOGLE_CLOUD_PROJECT=${projectId},GOOGLE_CLOUD_LOCATION=${region},GOOGLE_AGENT_ENGINE_ID=${agent_engine_id},GOOGLE_AGENT_ENGINE_LOCATION=${agent_engine_location},GEMINI_MODEL_VERSION="gemini-live-2.5-flash-preview-native-audio",COACH_BACKEND_URL=http://<YOUR_BACKEND_URL>

# --- MODIFIÉ : Déploiement du service frontend sur App Engine ---
# Assurez-vous d'avoir un fichier app.yaml dans votre répertoire ./front
echo "Deploying frontend service to App Engine..."
echo "first copy the front folder in a build dir"

mkdir build
cp -r ./front/* build/

echo "update the js files"
sed -i 's|const coachBackendURL = "http://localhost:8080"|const coachBackendURL = ""|g' build/js/app.js
sed -i 's|const coachAgentURL = "http://localhost:8081"|const coachAgentURL = ""|g' build/js/app.js

cd build
gcloud app deploy app.yaml --project=${projectId} --quiet
cd ..

echo "remove the build dir"
rm -rf build

echo "Setting up the Global External HTTPS Load Balancer..."

# Create endpoint NEG for the 2 Cloud Run services
gcloud compute network-endpoint-groups create coach-backend-neg \
  --project=${projectId} \
  --region=${region} \
  --network-endpoint-type=serverless \
  --cloud-run-service=coach-backend

gcloud compute network-endpoint-groups create coach-agent-wrapper-neg \
  --project=${projectId} \
  --region=${region} \
  --network-endpoint-type=serverless \
  --cloud-run-service=coach-agent-wrapper

gcloud compute network-endpoint-groups create coach-frontend-neg \
  --project=${projectId} \
  --region=${region} \
  --network-endpoint-type=serverless \
  --app-engine-service=coach

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
  --network-endpoint-group-region=${region}

gcloud compute backend-services add-backend coach-agent-wrapper-service \
  --project=${projectId} \
  --global \
  --network-endpoint-group=coach-agent-wrapper-neg \
  --network-endpoint-group-region=${region}

gcloud compute backend-services add-backend coach-frontend-service \
  --project=${projectId} \
  --global \
  --network-endpoint-group=coach-frontend-neg \
  --network-endpoint-group-region=us-central1
#  --network-endpoint-group-region=${region}

# Create a URL map to route requests to the correct backend service
gcloud compute url-maps create coach-url-map \
  --project=${projectId} \
  --default-service coach-frontend-service

# Add path matchers for the backend and agent services
gcloud compute url-maps add-path-matcher coach-url-map \
  --project=${projectId} \
  --default-service coach-frontend-service \
  --path-matcher-name=path-matcher-coach \
  --path-rules="/api/v1/daily-plans/*=coach-backend-service,/api/v1/categories/*=coach-backend-service,/api/v1/drills/*=coach-backend-service,/api/v1/plan-templates/*=coach-backend-service,/health=coach-backend-service,/api/v1/ws/*=coach-backend-service,/api/v1/chat/*=coach-agent-wrapper-service"

# --- HTTPS Configuration ---
# Le reste du script pour le Load Balancer, le certificat SSL et la redirection est inchangé.

echo "Configuring HTTPS..."

# Reserve a global static IP address for the load balancer
gcloud compute addresses create coach-lb-ip \
    --project=${projectId} \
    --ip-version=IPV4 \
    --global

# Create a Google-managed SSL certificate
# Note: Your domain must be verified in your Google Cloud project.
# This can take up to 60 minutes to provision.
echo "Creating Google-managed SSL certificate for ${domainName}..."
gcloud compute ssl-certificates create coach-ssl-cert \
  --project=${projectId} \
  --domains=${domainName} \
  --global

# Create a target HTTPS proxy to route requests to the URL map
gcloud compute target-https-proxies create coach-https-proxy \
  --project=${projectId} \
  --url-map=coach-url-map \
  --ssl-certificates=coach-ssl-cert \
  --global

# Create a global forwarding rule to route incoming requests to the HTTPS proxy
gcloud compute forwarding-rules create coach-https-rule \
  --project=${projectId} \
  --address=coach-lb-ip \
  --target-https-proxy=coach-https-proxy \
  --global \
  --ports=443

# --- HTTP-to-HTTPS Redirect ---
echo "Setting up HTTP-to-HTTPS redirect..."

# Create a URL map to handle the redirect
gcloud compute url-maps create coach-redirect-map \
  --project=${projectId} \
  --default-url-redirect "httpsRedirect=true,redirectResponseCode=MOVED_PERMANENTLY_DEFAULT"

# Create a target HTTP proxy for the redirect
gcloud compute target-http-proxies create coach-http-redirect-proxy \
  --project=${projectId} \
  --url-map coach-redirect-map

# Create a forwarding rule for HTTP traffic
gcloud compute forwarding-rules create coach-http-rule \
  --project=${projectId} \
  --address=coach-lb-ip \
  --global \
  --target-http-proxy=coach-http-redirect-proxy \
  --ports=80

# --- Final Instructions ---
echo "Deployment script finished."
echo "--------------------------"
echo "IMPORTANT: The SSL Certificate can take up to 60 minutes to be provisioned."
echo "You can check its status with: gcloud compute ssl-certificates describe coach-ssl-cert --global --project=${projectId}"
echo ""
echo "Your load balancer IP address is:"
gcloud compute addresses describe coach-lb-ip --project=${projectId} --global --format="value(address)"
echo ""
echo "Please create a DNS 'A' record for '${domainName}' pointing to this IP address."
