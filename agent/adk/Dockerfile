# Utiliser une image Python officielle comme image de base
FROM python:3.11-slim

# Définir le répertoire de travail dans le conteneur
WORKDIR /app

# Copier le fichier des dépendances (requirements.txt) en premier
# Cela permet de tirer parti de la mise en cache des couches Docker si les dépendances ne changent pas
COPY requirements.txt requirements.txt

# Installer les dépendances Python
RUN pip install --no-cache-dir -r requirements.txt

# Copier le reste du code de l'application dans le répertoire de travail
COPY coach_agent/* coach_agent/
COPY ../__init__.py .

# Exposer le port sur lequel votre application s'exécute (si c'est une application web)
# Remplacez 5000 par le port que votre application utilise
EXPOSE 8000

# Définir la commande pour exécuter votre application
# Remplacez app.py par le nom de votre fichier principal ou le module à exécuter
CMD ["adk", "api_server","--host","0.0.0.0"]