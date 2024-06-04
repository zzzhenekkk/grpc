package main

import (
	"context"
	"google.golang.org/grpc"
	"log"

	pb "src/transmitter"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTransmitterClient(conn)

	req := &pb.FrequencyRequest{ClientId: "example-client"}

	stream, err := client.TransmitFrequencies(context.Background(), req)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Fatalf("error while receiving: %v", err)
		}
		log.Printf("Session ID: %s, Frequency: %f, Timestamp: %d", resp.SessionId, resp.Frequency, resp.Timestamp)
	}
}
