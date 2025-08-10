from __future__ import annotations

import asyncio
import logging
from typing import Optional, Dict, Any

from typing_extensions import override
from typing_extensions import override

from google.adk.sessions import VertexAiSessionService
from google.adk.events.event import Event
from google.adk.sessions.session import Session

logger = logging.getLogger('google_adk.' + __name__)
# Set the logger to debug
# logger.setLevel(logging.DEBUG)



class VertexAiSessionServiceAsync(VertexAiSessionService):
    """Connects to the Vertex AI Agent Engine Session Service using GenAI API client.

    https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/sessions/overview
    """

    def __init__(self, project: Optional[str] = None, location: Optional[str] = None,
                 agent_engine_id: Optional[str] = None):
        """Initializes the VertexAiSessionService.

        Args:
          project: The project id of the project to use.
          location: The location of the project to use.
          agent_engine_id: The resource ID of the agent engine to use.
        """
        super().__init__(project, location, agent_engine_id)
        #  Queue and background task for event processing
        self._event_queue = asyncio.Queue()
        self._consumer_task: Optional[asyncio.Task] = None

    # Method to start the background consumer
    async def start(self):
        """Starts the background event consumer task."""
        if self._consumer_task is None:
            self._consumer_task = asyncio.create_task(self._event_consumer())
            logger.info("Vertex AI event consumer started.")

    # Method to stop the background consumer gracefully
    async def stop(self):
        """Stops the background event consumer task."""
        if self._consumer_task and not self._consumer_task.done():
            logger.info("Stopping Vertex AI event consumer...")
            try:
                # Add a sentinel value to the queue to signal the consumer to stop
                await self._event_queue.put(None)
                # Wait for the queue to be fully processed
                await self._event_queue.join()
                # Wait for the consumer task to finish
                await self._consumer_task
            except asyncio.CancelledError:
                logger.warning("Event consumer task was cancelled before completion.")
            finally:
                self._consumer_task = None
                logger.info("Vertex AI event consumer stopped.")

    # The background worker that consumes events from the queue
    async def _event_consumer(self):
        """Consumes events from the queue and sends them to the reasoning engine."""
        while True:
            item = await self._event_queue.get()
            if item is None:
                # Sentinel value received, stop the loop
                self._event_queue.task_done()
                break

            session, event = item
            try:
                #Skip the audio event if it comes from the agent itself, keep only the text event from it. Keep the audio from the user
                if event.content is None:
                    continue
                if event.content and event.content.parts and not event.content.parts[0].text and event.content.parts[0].inline_data and event.content.role == "model":
                    continue
                logger.debug(f"Processing event {event.id} for session {session.id} from queue with text {event.content.parts[0].text}.")

                reasoning_engine_id = self._get_reasoning_engine_id(session.app_name)
                # It's important to get a fresh api_client in case of token expiry, etc.
                api_client = self._get_api_client()
                await api_client.async_request(
                    http_method='POST',
                    path=f'reasoningEngines/{reasoning_engine_id}/sessions/{session.id}:appendEvent',
                    request_dict=_convert_event_to_json(event),
                )
                logger.debug(f"Successfully sent event {event.id} for session {session.id}.")
            except Exception as e:
                logger.error(f"Error sending event {event.id} for session {session.id}: {e}")
            finally:
                # Signal that the task is done, regardless of success or failure
                self._event_queue.task_done()

    #  append_event is now non-blocking
    @override
    async def append_event(self, session: Session, event: Event) -> Event:
        # Update the in-memory session. This is fast and remains awaited.
        # I must avoid the parent call that create latency and jump to "grand parent" call instead.
        await super(VertexAiSessionService, self).append_event(session=session, event=event)

        # Instead of making the API call directly, put the event in the queue
        # to be processed by the background task. This is non-blocking.
        await self._event_queue.put((session, event))
        logger.debug(f"Queued event {event.id} for session {session.id}")

        return event

# Copy private method for simplicity
def _convert_event_to_json(event: Event) -> Dict[str, Any]:
    metadata_json = {
        'partial': event.partial,
        'turn_complete': event.turn_complete,
        'interrupted': event.interrupted,
        'branch': event.branch,
        'custom_metadata': event.custom_metadata,
        'long_running_tool_ids': (
            list(event.long_running_tool_ids)
            if event.long_running_tool_ids
            else None
        ),
    }
    if event.grounding_metadata:
        metadata_json['grounding_metadata'] = event.grounding_metadata.model_dump(
            exclude_none=True, mode='json'
        )

    event_json = {
        'author': event.author,
        'invocation_id': event.invocation_id,
        'timestamp': {
            'seconds': int(event.timestamp),
            'nanos': int(
                (event.timestamp - int(event.timestamp)) * 1_000_000_000
            ),
        },
        'error_code': event.error_code,
        'error_message': event.error_message,
        'event_metadata': metadata_json,
    }

    if event.actions:
        actions_json = {
            'skip_summarization': event.actions.skip_summarization,
            'state_delta': event.actions.state_delta,
            'artifact_delta': event.actions.artifact_delta,
            'transfer_agent': event.actions.transfer_to_agent,
            'escalate': event.actions.escalate,
            'requested_auth_configs': event.actions.requested_auth_configs,
        }
        event_json['actions'] = actions_json
    if event.content:
        event_json['content'] = event.content.model_dump(
            exclude_none=True, mode='json'
        )
    if event.error_code:
        event_json['error_code'] = event.error_code
    if event.error_message:
        event_json['error_message'] = event.error_message
    return event_json

