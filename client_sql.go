package main

import (
	"context"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc"
	"log"
	"math"
	pb "src/transmitter"
	"sync"
)

type AnomalyDetector struct {
	mu        sync.Mutex
	n         int
	mean      float64
	m2        float64
	stdDev    float64
	anomalies []pb.FrequencyResponse
}

func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		n:         0,
		mean:      0,
		m2:        0,
		stdDev:    0,
		anomalies: []pb.FrequencyResponse{},
	}
}

func (ad *AnomalyDetector) AddData(frequency float64) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	ad.n++
	delta := frequency - ad.mean
	ad.mean += delta / float64(ad.n)
	delta2 := frequency - ad.mean
	ad.m2 += delta * delta2

	if ad.n > 1 {
		ad.stdDev = math.Sqrt(ad.m2 / float64(ad.n-1))
	}
}

func (ad *AnomalyDetector) CheckAnomaly(frequency float64, k float64) bool {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	if math.Abs(frequency-ad.mean) > k*ad.stdDev {
		return true
	}
	return false
}

type Anomaly struct {
	SessionID string
	Frequency float64
	Timestamp int64
}

func SaveAnomaly(conn *pgx.Conn, anomaly Anomaly) error {
	_, err := conn.Exec(context.Background(), "INSERT INTO anomalies (session_id, frequency, timestamp) VALUES ($1, $2, $3)", anomaly.SessionID, anomaly.Frequency, anomaly.Timestamp)
	return err
}

func main() {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:1234@localhost:5432/anomaly_db")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	grpcConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer grpcConn.Close()

	client := pb.NewTransmitterClient(grpcConn)
	req := &pb.FrequencyRequest{ClientId: "anomaly-detector"}

	stream, err := client.TransmitFrequencies(context.Background(), req)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	detector := NewAnomalyDetector()
	var k float64 = 2.0

	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Fatalf("error while receiving: %v", err)
		}

		detector.AddData(resp.Frequency)
		if detector.n > 50 {
			if detector.CheckAnomaly(resp.Frequency, k) {
				log.Printf("Anomaly detected: %v", resp)
				anomaly := Anomaly{
					SessionID: resp.SessionId,
					Frequency: resp.Frequency,
					Timestamp: resp.Timestamp,
				}
				err := SaveAnomaly(conn, anomaly)
				if err != nil {
					log.Printf("Failed to save anomaly: %v", err)
				}
				detector.anomalies = append(detector.anomalies, *resp)
			}
		}

		log.Printf("Processed value: %f, mean: %f, stdDev: %f", resp.Frequency, detector.mean, detector.stdDev)
	}
}
