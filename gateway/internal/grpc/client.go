package grpcserver

import (
	"context"
	"fmt"
	"io"

	pb "github.com/hotel-voice-agent/gateway/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AgentClient struct {
	client pb.AgentServiceClient
	conn   *grpc.ClientConn
}

func NewAgentClient(addr string) (*AgentClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to python brain: %w", err)
	}
	
	client := pb.NewAgentServiceClient(conn)
	return &AgentClient{
		client: client,
		conn:   conn,
	}, nil
}

// StreamTranscript sends a transcript to the Python Brain and invokes the chunkHandler for each chunk of text received
func (ac *AgentClient) StreamTranscript(ctx context.Context, sessionID, transcript string, chunkHandler func(string)) error {
	req := &pb.TranscriptRequest{
		SessionId:  sessionID,
		Transcript: transcript,
	}

	stream, err := ac.client.ProcessTranscript(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to call ProcessTranscript: %w", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error receiving from ProcessTranscript stream: %w", err)
		}

		if resp.TextChunk != "" {
			chunkHandler(resp.TextChunk)
		}
		
		if resp.IsDone {
			break
		}
	}

	return nil
}

func (ac *AgentClient) Close() error {
	if ac.conn != nil {
		return ac.conn.Close()
	}
	return nil
}
