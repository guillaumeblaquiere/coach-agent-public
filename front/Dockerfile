# Utiliser une image NGINX officielle comme image de base
FROM nginx:alpine

# Copier votre fichier HTML dans le répertoire par défaut de NGINX pour le contenu web
# Remplacez 'index.html' par le nom de votre fichier HTML s'il est différent.
# Assurez-vous que votre fichier HTML est dans le même répertoire que ce Dockerfile,
# ou ajustez le chemin source en conséquence.
COPY index.html /usr/share/nginx/html/index.html

# set the const URL to "" in the file
RUN sed -i 's|const coachBackendURL = "http://localhost:8080"|const coachBackendURL = ""|g' /usr/share/nginx/html/index.html && \
sed -i 's|const coachAgentURL = "http://localhost:8081"|const coachAgentURL = ""|g' /usr/share/nginx/html/index.html

# NGINX expose le port 80 par défaut, donc pas besoin de le déclarer explicitement ici
# à moins que vous ne vouliez le documenter ou le changer.
#EXPOSE 8080

# La commande par défaut de l'image NGINX démarrera le serveur NGINX.