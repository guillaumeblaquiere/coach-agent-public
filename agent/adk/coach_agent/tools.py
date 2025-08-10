import asyncio
import datetime
import os
from typing import AsyncGenerator
from urllib.parse import urljoin

import requests

sourceRequest = "agent"

def get_date_today() -> str:
    """Returns today's date in YYYY-MM-DD format."""
    return datetime.date.today().strftime("%Y-%m-%d")


def get_achievement(date: str) -> dict:
    """
    Retrieves the achievement  for a specified date.

    Args:
        date (str): The date for which to retrieve the training plan in YYYY-MM-DD format or "today" for today.

    :return: the training plan for the day organized by categories and drills.
    """
    coach_backend_url = os.getenv("COACH_BACKEND_URL")
    if not coach_backend_url:
        return {"status": "error", "error_message": "COACH_BACKEND_URL environment variable is not set."}

    if date == "today":
        url = f"{coach_backend_url}/api/v1/daily-plans/today"
    else:
        url = f"{coach_backend_url}/api/v1/daily-plans/{date}"

    try:
        # Follow the redirect
        response = requests.get(url, allow_redirects=True)
        response.raise_for_status()  # Raise an exception for bad status codes
        return {"status": "success", "report": response.json()}
    except requests.exceptions.RequestException as e:
        return {"status": "error", "error_message": f"Failed to retrieve training plan: {e}"}


def get_training_template(template_id: str) -> dict:
    """
    Retrieves the training plan template for a specified template ID.

    Args:
        template_id (str): The ID of the training plan template.

    :return: the training plan template organized by categories and drills, with the descriptions and the target number of repetitions to complete the drill.
    """
    coach_backend_url = os.getenv("COACH_BACKEND_URL")
    if not coach_backend_url:
        return {"status": "error", "error_message": "COACH_BACKEND_URL environment variable is not set."}

    url = f"{coach_backend_url}/api/v1/plan-templates/{template_id}"

    try:
        response = requests.get(url)
        response.raise_for_status()  # Raise an exception for bad status codes
        return {"status": "success","report": response.json()}
    except requests.exceptions.RequestException as e:
        return {"status": "error", "error_message": f"Failed to retrieve training template: {e}"}


def update_achievement(plan: dict) -> dict:
    """
    Updates the achievement for today. The minimal JSON format is
    {
      "id": "<ID of the plan>",
      "date": "<DATE format YYYY-MM-DD>",
      "repetitions": {
        "<CATEGORY_ID>": {
          "<DRILL_ID>": {
            "repetition": <OPTIONAL_NUMBER_OF_REPETITIONS>,
            "note": "<OPTIONAL_NOTE>"
          },
          ...
        },
        ...
      }
    }
    Only the ID are used (category and drill)
    If a whole category must be updated, all the drills of the category can be added in the plan to update
    The plan can be updated with:
     - a new repetition value on the drill
     - a new note on the drill
     - both a repetition value and note on the drill

    Args:
        plan (dict): The updated training plan data.

    :return: The status of the API call and
     - either the updated training plan from the backend.
     - or an error message.
    """
    coach_backend_url = os.getenv("COACH_BACKEND_URL")
    if not coach_backend_url:
        return {"status": "error", "error_message": "COACH_BACKEND_URL environment variable is not set."}

    current_url = f"{coach_backend_url}/api/v1/daily-plans/today?source={sourceRequest}"
    max_redirects = 5  # Limite pour éviter les boucles infinies

    try:
        for redirect_attempt in range(max_redirects):
            response = requests.put(current_url, json=plan, allow_redirects=False)

            # Vérifier si c'est une redirection que nous devons suivre manuellement
            if response.status_code in [301, 302, 303, 307, 308] and 'Location' in response.headers:
                redirect_url = response.headers['Location']

                # Construire une URL absolue si la redirection est relative
                if not redirect_url.startswith(('http://', 'https://')):
                    redirect_url = urljoin(current_url, redirect_url)

                current_url = redirect_url

                # Pour 301, 302, 303, requests changerait la méthode en GET.
                # En continuant la boucle, nous ré-émettons un PUT.
                # Pour 307, 308, la méthode serait préservée par requests,
                # mais notre gestion manuelle le fait aussi.
                if redirect_attempt == max_redirects - 1:
                    # Limite de redirection atteinte
                    return {"status": "error", "error_message": "Too many redirects."}
                # Continue la boucle pour faire la requête PUT vers la nouvelle URL
            else:
                # Pas une redirection que nous gérons, ou pas de redirection du tout
                response.raise_for_status()  # Lève une exception pour les codes d'erreur HTTP
                return {"status": "success", "report": response.json()}

        # Si la boucle se termine sans retourner (ne devrait pas arriver avec la logique ci-dessus)
        return {"status": "error", "error_message": "Redirect handling failed unexpectedly."}
    except requests.exceptions.RequestException as e:
        return {"status": "error", "error_message": f"Failed to update training plan: {e}"}

# async def start_timer(duration_in_seconds: int) -> AsyncGenerator[str, None]:
async def start_timer(duration_in_seconds: int) -> str:
    """
    Starts a timer for a specified duration.
    It first yields a confirmation message, waits for the duration without blocking,
    and then yields a completion message.

    Args:
        duration_in_seconds (int): The duration of the timer in seconds.
    """
    if duration_in_seconds <= 0:
        yield "La durée doit être supérieure à zéro."
        return
    print(f"Starting timer for {duration_in_seconds} seconds")
    yield "C'est parti !"
    await asyncio.sleep(duration_in_seconds)
    print("Timer finished")
    yield "Terminé !"
