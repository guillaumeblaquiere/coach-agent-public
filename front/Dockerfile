# Use an official NGINX image as a base image
FROM nginx:alpine

# Copy the static assets to the default NGINX web content directory.
# We now copy the index.html file and the css and js directories.
COPY index.html /usr/share/nginx/html/
COPY css/ /usr/share/nginx/html/css/
COPY js/ /usr/share/nginx/html/js/

# In the container, set the backend URLs to be relative paths ("").
# This is crucial for the reverse proxy setup (e.g., in docker-compose)
# to correctly route API requests from the browser.
# The target file is now js/app.js where the constants are defined.
RUN sed -i 's|const coachBackendURL = "http://localhost:8080"|const coachBackendURL = ""|g' /usr/share/nginx/html/js/app.js && \
    sed -i 's|const coachAgentURL = "http://localhost:8081"|const coachAgentURL = ""|g' /usr/share/nginx/html/js/app.js

# NGINX exposes port 80 by default, so no need to explicitly declare it
# unless you want to document or change it.
# EXPOSE 80

# The default command of the NGINX image will start the NGINX server.