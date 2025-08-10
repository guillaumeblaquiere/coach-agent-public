import os
from google.adk.agents import Agent
from google.adk.planners import BuiltInPlanner

from google.genai.types import  ThinkingConfig


from . import tools
from .prompt import get_prompt


root_agent = Agent(
    name="coach_agent",
    model=os.getenv("GEMINI_MODEL_VERSION"),
    description=(
        """
        Tu es un agent qui coach l'utilisateur sur des seances de stretching. 
        """
    ),
    # instruction=get_prompt["direct"],
    instruction=get_prompt["english_direct"],
    tools=[
        tools.get_achievement,
        tools.update_achievement,
        tools.get_training_template,
        tools.get_date_today,
        tools.start_timer,
    ],
    planner=BuiltInPlanner(
        thinking_config=ThinkingConfig(
            include_thoughts=True,
            thinking_budget=1024,
        )
    ),
)
