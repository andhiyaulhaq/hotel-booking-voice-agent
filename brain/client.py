import grpc
import sys
from pathlib import Path

# Add proto generated files to path
proto_dir = str(Path(__file__).parent / "src" / "proto")
sys.path.append(proto_dir)

import service_pb2
import service_pb2_grpc

def run():
    print("Connecting to local Python gRPC Brain on port 50052...")
    try:
        with grpc.insecure_channel('localhost:50052') as channel:
            stub = service_pb2_grpc.AgentServiceStub(channel)
            
            transcript = "I want to book a suite for 2 nights. My name is Alice."
            print(f"Sending transcript: '{transcript}'")
            
            request = service_pb2.TranscriptRequest(
                session_id="test-session-123",
                transcript=transcript
            )
            
            print("Response chunks from Agent:")
            for chunk in stub.ProcessTranscript(request):
                if chunk.text_chunk:
                    print(chunk.text_chunk, end="", flush=True)
                if chunk.is_done:
                    print("\n\n[Stream Completed]")
                    break
                    
    except grpc.RpcError as e:
        print(f"\nRPC Error: {e.details()}")
    except Exception as e:
        print(f"\nError: {e}")

if __name__ == '__main__':
    run()
