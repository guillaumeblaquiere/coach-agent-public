import requests
import datetime
import os
from zoneinfo import ZoneInfo
from google.adk.agents import Agent
from urllib.parse import urljoin

from .prompt import get_prompt


def get_weather(city: str) -> dict:
    """Retrieves the current weather report for a specified city.

    Args:
        city (str): The name of the city for which to retrieve the weather report.

    Returns:
        dict: status and result or error msg.
    """
    if city.lower() == "new york":
        return {
            "status": "success",
            "report": (
                "The weather in New York is sunny with a temperature of 25 degrees"
                " Celsius (77 degrees Fahrenheit)."
            ),
        }
    else:
        return {
            "status": "error",
            "error_message": f"Weather information for '{city}' is not available.",
        }


def get_current_time(city: str) -> dict:
    """Returns the current time in a specified city.

    Args:
        city (str): The name of the city for which to retrieve the current time.

    Returns:
        dict: status and result or error msg.
    """

    if city.lower() == "new york":
        tz_identifier = "America/New_York"
    else:
        return {
            "status": "error",
            "error_message": (
                f"Sorry, I don't have timezone information for {city}."
            ),
        }

    tz = ZoneInfo(tz_identifier)
    now = datetime.datetime.now(tz)
    report = (
        f'The current time in {city} is {now.strftime("%Y-%m-%d %H:%M:%S %Z%z")}'
    )
    return {"status": "success", "report": report}


def cheer_up():
    """Provides a cheering message."""
    return {"status": "success", "report": "You're doing great! Keep up the good work!"}

def get_date_today() -> str:
    """Returns today's date in YYYY-MM-DD format."""
    return datetime.date.today().strftime("%Y-%m-%d")


def get_training_plan(date: str) -> dict:
    """
    Retrieves the daily training plan for a specified date.

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
        return {"status": "success", "actionDone": "GET", "report": response.json()}
    except requests.exceptions.RequestException as e:
        return {"status": "error", "error_message": f"Failed to retrieve training plan: {e}"}

def update_training(plan: dict) -> dict:
    """
    Updates the daily training plan for today. The minimal JSON format is
    {
      "id": "<ID of the plan>",
      "date": "<DATE format YYYY-MM-DD>",
      "categories": {
        "<category ID>": {
          "drills": {
            "<drill ID>": {
              "repetition": <value to update>
            }
            "<optional other drill ID and repetition>"
          }
        }
      }
    }
    If a whole category must be updated, all the drills of the category can be added in the plan to update

    Args:
        plan (dict): The updated training plan data.

    :return: The status of the API call and
     - either the updated training plan from the backend.
     - or an error message.
    """
    coach_backend_url = os.getenv("COACH_BACKEND_URL")
    if not coach_backend_url:
        return {"status": "error", "error_message": "COACH_BACKEND_URL environment variable is not set."}

    current_url = f"{coach_backend_url}/api/v1/daily-plans/today"
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
                return {"status": "success", "actionDone": "PUT", "report": response.json()}

        # Si la boucle se termine sans retourner (ne devrait pas arriver avec la logique ci-dessus)
        return {"status": "error", "error_message": "Redirect handling failed unexpectedly."}
    except requests.exceptions.RequestException as e:
        return {"status": "error", "error_message": f"Failed to update training plan: {e}"}



root_agent = Agent(
    name="coach_agent",
    model="gemini-2.0-flash",
    description=(
        """
        Tu es un agent qui coach l'utilisateur sur des seances de stretching. 
        """
    ),
    # instruction=get_prompt["exp"],
    instruction=get_prompt["working"],
    tools=[get_training_plan, get_date_today,update_training],

)

