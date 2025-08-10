import os
import json
import asyncio
import base64
from contextlib import asynccontextmanager

from typing import List, Literal, Optional, Any
import logging

from fastapi.logger import logger
from google.adk.artifacts import InMemoryArtifactService
from google.adk.auth.credential_service.in_memory_credential_service import InMemoryCredentialService
from google.adk.cli.cli_eval import EVAL_SESSION_ID_PREFIX
from google.adk.cli.utils import envs, cleanup
from google.adk.memory import VertexAiMemoryBankService
from google.adk.sessions import Session, VertexAiSessionService

from google.genai.types import (
    Part,
    Content,
    Blob, AudioTranscriptionConfig,
)

from google.adk.runners import Runner
from google.adk.agents import LiveRequestQueue
from google.adk.agents.run_config import RunConfig

from fastapi import FastAPI, WebSocket, Query, HTTPException
from starlette.middleware.cors import CORSMiddleware
from starlette.types import Lifespan
from starlette.websockets import WebSocketDisconnect

from coach_agent.tools import sourceRequest
from coach_agent.agent import root_agent
from vertex_ai_session_service_async import VertexAiSessionServiceAsync

logging.basicConfig(level=logging.INFO)
logger.setLevel(logging.INFO)

allowed_origins = [
    "http://localhost:8080", # Assuming your Go backend serves the frontend here
    "http://127.0.0.1:8080",
    "http://localhost:63342", # Common for WebStorm's built-in server
    # Add other origins if your frontend is hosted elsewhere
]


def createSessionService():
    # return InMemorySessionService()

    project_id = os.getenv("GOOGLE_CLOUD_PROJECT")
    location = os.getenv("GOOGLE_AGENT_ENGINE_LOCATION")
    agent_engine_id = os.getenv("GOOGLE_AGENT_ENGINE_ID")

    service = VertexAiSessionServiceAsync(
        project=project_id,
        location=location,
        agent_engine_id=agent_engine_id
    )

    # Original session service
    # service = VertexAiSessionService(
    #     project=project_id,
    #     location=location,
    #     agent_engine_id=agent_engine_id
    # )

    return service


def createMemoryService():
    # return InMemoryMemoryService()

    project_id = os.getenv("GOOGLE_CLOUD_PROJECT")
    location = os.getenv("GOOGLE_AGENT_ENGINE_LOCATION")
    agent_engine_id = os.getenv("GOOGLE_AGENT_ENGINE_ID")

    service = VertexAiMemoryBankService(
        project=project_id,
        location=location,
        agent_engine_id=agent_engine_id
    )
    return service


def get_my_app(
        *,
        agents_dir: str,
        session_service_uri: Optional[str] = None,
        artifact_service_uri: Optional[str] = None,
        memory_service_uri: Optional[str] = None,
        allow_origins: Optional[list[str]] = None,
        lifespan: Optional[Lifespan[FastAPI]] = None,
) -> FastAPI:

    # session_service is created once per app instance, which is correct for a shared service
    session_service = createSessionService()
    memory_service = createMemoryService()
    artifact_service = InMemoryArtifactService()
    # Build  the Credential service
    credential_service = InMemoryCredentialService()

    @asynccontextmanager
    async def internal_lifespan(app: FastAPI):
        # Initialize app state
        app.state.runners = {}
        app.state.active_websockets = {}

        # Initialize session service worker
        if isinstance(session_service, VertexAiSessionServiceAsync):
            await session_service.start()

        try:
            if lifespan:
                async with lifespan(app) as lifespan_context:
                    yield lifespan_context
            else:
                yield
        finally:
            # Stop session service worker
            if isinstance(session_service, VertexAiSessionService):
                await session_service.stop()

            # Cleanup runners
            logger.info("Closing all cached runners...")
            await cleanup.close_runners(list(app.state.runners.values()))
            app.state.runners.clear()
            logger.info("All cached runners closed.")

    app = FastAPI(lifespan=internal_lifespan)
    if allow_origins:
        app.add_middleware(
            CORSMiddleware,
            allow_origins=allow_origins,
            allow_credentials=True,
            allow_methods=["*"],
            allow_headers=["*"],
        )


    # create health check for Load Balancer and Cloud Run sidecar deployment
    @app.get("/health")
    async def health_check():
        return {"status": "ok"}


    #### Session management (No changes here)
    @app.get(
        "/apps/{app_name}/users/{user_id}/sessions/{session_id}",
        response_model_exclude_none=True,
    )
    async def get_session(
            app_name: str, user_id: str, session_id: str
    ) -> Session:
        session = await session_service.get_session(
            app_name=app_name, user_id=user_id, session_id=session_id
        )
        if not session:
            raise HTTPException(status_code=404, detail="Session not found")
        return session

    @app.get(
        "/apps/{app_name}/users/{user_id}/sessions",
        response_model_exclude_none=True,
    )
    async def list_sessions(app_name: str, user_id: str) -> list[Session]:
        list_sessions_response = await session_service.list_sessions(
            app_name=app_name, user_id=user_id
        )
        return [
            session
            for session in list_sessions_response.sessions
            # Remove sessions that were generated as a part of Eval.
            if not session.id.startswith(EVAL_SESSION_ID_PREFIX)
        ]

    @app.post(
        "/apps/{app_name}/users/{user_id}/agentEngineSessions",
        response_model_exclude_none=True,
    )
    async def create_session_with_agent_engine(
            app_name: str,
            user_id: str,
            state: Optional[dict[str, Any]] = None,
    ) -> Session:
        return await session_service.create_session(
            app_name=app_name, user_id=user_id, state=state
        )

    @app.delete("/apps/{app_name}/users/{user_id}/sessions/{session_id}")
    async def delete_session(app_name: str, user_id: str, session_id: str):
        await session_service.delete_session(
            app_name=app_name, user_id=user_id, session_id=session_id
        )


    ##### Websocket management
    @app.websocket("/run_live")
    async def agent_live_run(
            websocket: WebSocket,
            app_name: str,
            user_id: str,
            session_id: str,
            modalities: List[Literal["TEXT", "AUDIO"]] = Query(
                default=["TEXT", "AUDIO"]
            ),
    ) -> None:
        # --- NEW: Handle reconnection by closing the previous connection ---
        if session_id in app.state.active_websockets:
            logger.warning(
                f"Session {session_id} already has an active connection. "
                f"Closing the old one to establish a new one."
            )
            old_websocket, old_tasks = app.state.active_websockets[session_id]
            try:
                # Closing the websocket will trigger WebSocketDisconnect in the old handler,
                # which will then execute its `finally` block for cleanup.
                await old_websocket.close(code=1001, reason="New connection established")
            except Exception as e:
                # The old socket might already be dead, which is fine.
                logger.info(f"Could not close old websocket for session {session_id} (may already be closed): {e}")

            # Proactively cancel the old tasks as a safeguard.
            for task in old_tasks:
                task.cancel()
            await asyncio.gather(*old_tasks, return_exceptions=True)
            logger.info(f"Old connection tasks for session {session_id} cancelled.")
        # --- END NEW ---

        await websocket.accept()
        logger.info(f"User {user_id} connected to agent {app_name} with session {session_id} and modalities {modalities}")

        modality = "AUDIO" if "AUDIO" in modalities else "TEXT"
        run_config = RunConfig(
            response_modalities=[modality],
            output_audio_transcription=AudioTranscriptionConfig(),
            input_audio_transcription=AudioTranscriptionConfig(),
            # enable_affective_dialog=True,
            # proactivity=ProactivityConfig(),
        )

        if app_name not in app.state.runners:
            logger.info(f"Creating new Runner for app_name: {app_name} and caching it.")
            app.state.runners[app_name] = Runner(
                app_name=app_name,
                agent=root_agent,
                artifact_service=artifact_service,
                session_service=session_service,
                memory_service=memory_service,
                credential_service=credential_service,
            )
        runner = app.state.runners[app_name]

        live_request_queue = LiveRequestQueue()
        live_event_generator = runner.run_live(
            session_id=session_id, user_id=user_id, live_request_queue=live_request_queue, run_config=run_config
        )

        async def send_response():
            try:
                async for event in live_event_generator:
                    if event.turn_complete or event.interrupted:
                        message = {
                            "turn_complete": event.turn_complete,
                            "interrupted": event.interrupted,
                        }
                        await websocket.send_text(json.dumps(message))
                        print(f"Received event: {event}")
                        logger.info(f"[AGENT TO CLIENT]: {message}")
                        continue

                    part: Part = event.content and event.content.parts and event.content.parts[0]
                    if not part:
                        continue

                    is_audio = part.inline_data and part.inline_data.mime_type.startswith("audio/pcm")
                    if is_audio:
                        audio_data = part.inline_data and part.inline_data.data
                        if audio_data:
                            message = {"mime_type": "audio/pcm", "data": base64.b64encode(audio_data).decode("ascii")}
                            await websocket.send_text(json.dumps(message))
                            continue
                    if part.text and event.partial:
                        message = {"mime_type": "text/plain", "data": part.text}
                        await websocket.send_text(json.dumps(message))
                    else:
                        print(f"Received unhandled event: {event}")
            except asyncio.CancelledError:
                logger.info(f"send_response task for session {session_id} cancelled.")
            except Exception as e:
                logger.exception(f"Error in send_response task for session {session_id}: %s", e)
            finally:
                if hasattr(live_event_generator, 'aclose'):
                    await live_event_generator.aclose()
                logger.info(f"Live event generator for session {session_id} closed.")

        async def received_message():
            try:
                while True:
                    message_json = await websocket.receive_text()
                    message = json.loads(message_json)

                    if "mime_type" not in message or "data" not in message:
                        logger.warning("Received a message with an unexpected format, ignoring: %s", message)
                        continue

                    mime_type = message["mime_type"]
                    data = message["data"]

                    if mime_type == "text/plain":
                        content = Content(role="user", parts=[Part.from_text(text=data)])
                        live_request_queue.send_content(content=content)
                        logger.info(f"[CLIENT TO AGENT] ({session_id}): Text message: {data}")
                    elif mime_type == "audio/pcm":
                        decoded_data = base64.b64decode(data)
                        live_request_queue.send_realtime(Blob(data=decoded_data, mime_type=mime_type))
                    elif mime_type == "application/json":
                        event_source = message.get("event_source")
                        if event_source != sourceRequest:
                            event_type = message.get("event_type")
                            event_data = message.get("data")
                            logger.info(f"Received UI event '{event_type}' from wrapper.")
                            system_notification = (
                                f"System Notification: The user has just performed the action '{event_type}'.\n"
                                f"Here is the new state of their plan:\n"
                                f"{json.dumps(event_data, indent=2)}"
                            )
                            content = Content(role="user", parts=[Part.from_text(text=system_notification)])
                            live_request_queue.send_content(content=content)
                    else:
                        logger.warning("Mime type not supported: %s", mime_type)
            except WebSocketDisconnect:
                logger.info(f"Client for session {session_id} disconnected during received_message.")
            except asyncio.CancelledError:
                logger.info(f"received_message task for session {session_id} cancelled.")
            except json.JSONDecodeError as je:
                logger.error("Failed to decode JSON from client: %s", je)
            except Exception as e:
                logger.exception(f"Error in received_message task for session {session_id}: %s", e)

        tasks = [
            asyncio.create_task(send_response()),
            asyncio.create_task(received_message()),
        ]

        app.state.active_websockets[session_id] = (websocket, tasks)

        try:
            done, pending = await asyncio.wait(
                tasks, return_when=asyncio.FIRST_COMPLETED
            )
            for task in done:
                if task.exception():
                    raise task.exception()
        except WebSocketDisconnect:
            logger.info(f"Client for session {session_id} disconnected gracefully.")
        except Exception as e:
            logger.exception("Error during live websocket communication for session %s: %s", session_id, e)
            WEBSOCKET_INTERNAL_ERROR_CODE = 1011
            WEBSOCKET_MAX_BYTES_FOR_REASON = 123
            await websocket.close(
                code=WEBSOCKET_INTERNAL_ERROR_CODE,
                reason=str(e)[:WEBSOCKET_MAX_BYTES_FOR_REASON],
            )
        finally:
            logger.info(f"Cleaning up resources for session {session_id}.")

            # Remove the session from the active list, but only if it's this specific instance.
            # This prevents a race condition where a new connection might have already been registered.
            if session_id in app.state.active_websockets and app.state.active_websockets[session_id][0] is websocket:
                del app.state.active_websockets[session_id]
                logger.info(f"Session {session_id} removed from active connections.")

            # Cancel all tasks associated with this connection.
            for task in tasks:
                task.cancel()

            # Wait for tasks to finish their cancellation sequence.
            await asyncio.gather(*tasks, return_exceptions=True)

            # The generator is closed by send_response's `finally` block.
            # The queue must be closed here.
            live_request_queue.close()
            logger.info(f"Live request queue for session {session_id} closed.")

    return app

# Instantiate the FastAPI application
app = get_my_app(
    agents_dir=".",
    allow_origins=allowed_origins
)