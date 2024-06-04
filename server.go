package main

import (
	"log"
	"math/rand"
	"net"
	"time"

	pb "src/transmitter"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTransmitterServer
}

// TransmitFrequencies реализует серверный поток gRPC
func (s *server) TransmitFrequencies(req *pb.FrequencyRequest, stream pb.Transmitter_TransmitFrequenciesServer) error {
	sessionID := uuid.New().String()
	mean := -10 + rand.Float64()*(20)    // Случайное значение mean от -10 до 10
	stdDev := 0.3 + rand.Float64()*(1.2) // Случайное значение stdDev от 0.3 до 1.5

	log.Printf("New session: %s, mean: %f, stdDev: %f", sessionID, mean, stdDev)

	for {
		frequency := rand.NormFloat64()*stdDev + mean
		timestamp := time.Now().UTC().Unix()

		resp := &pb.FrequencyResponse{
			SessionId: sessionID,
			Frequency: frequency,
			Timestamp: timestamp,
		}
		if err := stream.Send(resp); err != nil {
			return err
		}

		time.Sleep(time.Second)
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTransmitterServer(s, &server{})

	log.Printf("Server is listening on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
