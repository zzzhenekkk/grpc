package main

import (
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	pb "path/to/your/generated/protobuf/package"
)

type server struct {
	pb.UnimplementedTransmitterServer
}

func (s *server) TransmitFrequencies(req *pb.FrequencyRequest, stream pb.Transmitter_TransmitFrequenciesServer) error {
	// Генерация нового session_id и случайных значений для mean и stdDev
	sessionID := uuid.New().String()
	mean := -10 + rand.Float64()*(10-(-10))
	stdDev := 0.3 + rand.Float64()*(1.5-0.3)

	log.Printf("New session: %s, mean: %f, stdDev: %f", sessionID, mean, stdDev)

	for {
		// Генерация случайной частоты на основе нормального распределения
		frequency := rand.NormFloat64()*stdDev + mean
		timestamp := time.Now().UTC().Unix()

		// Отправка ответа клиенту
		resp := &pb.FrequencyResponse{
			SessionId: sessionID,
			Frequency: frequency,
			Timestamp: timestamp,
		}
		if err := stream.Send(resp); err != nil {
			return err
		}

		// Задержка перед отправкой следующего сообщения
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
