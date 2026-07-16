import grpc
from concurrent import futures
import sys
from pathlib import Path
import logging
from langchain_core.messages import HumanMessage

# Setup paths for proto imports
proto_dir = str(Path(__file__).parent / "proto")
sys.path.append(proto_dir)

import service_pb2
import service_pb2_grpc
from agent import create_concierge_agent

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class AgentServiceServicer(service_pb2_grpc.AgentServiceServicer):
    def __init__(self):
        self.agent = create_concierge_agent()

    def ProcessTranscript(self, request, context):
        """
        Receives a transcript from Go, processes it through LangChain,
        and streams the AI's response chunks back to Go.
        """
        logger.info(f"Received transcript from {request.session_id}: {request.transcript}")
        
        # We pass the session_id as the thread_id to maintain conversation history per user
        config = {"configurable": {"thread_id": request.session_id}}
        
        # We stream the messages from the LangChain agent
        try:
            stream = self.agent.stream(
                {"messages": [HumanMessage(content=request.transcript)]},
                config,
                stream_mode="messages"
            )
            
            for message, metadata in stream:
                if hasattr(message, 'content') and isinstance(message.content, str) and message.content:
                    # Stream text chunks back to Go
                    yield service_pb2.AgentResponse(
                        text_chunk=message.content,
                        is_done=False
                    )
                    
        except Exception as e:
            logger.error(f"Error processing transcript: {e}")
            yield service_pb2.AgentResponse(
                text_chunk="I'm sorry, I encountered an internal error processing that request.",
                is_done=False
            )
            
        # Signal completion
        yield service_pb2.AgentResponse(
            text_chunk="",
            is_done=True
        )

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_AgentServiceServicer_to_server(AgentServiceServicer(), server)
    # The Python brain listens on port 50052
    server.add_insecure_port('[::]:50052')
    logger.info("Python Agent Brain starting on port 50052...")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
