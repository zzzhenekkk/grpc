package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"math"
	pb "src/transmitter"
	"sync"
)

// AnomalyDetector struct to hold mean, stdDev and other necessary fields
type AnomalyDetector struct {
	mu        sync.Mutex
	n         int
	mean      float64
	m2        float64
	stdDev    float64
	anomalies []pb.FrequencyResponse
}

// NewAnomalyDetector returns a new AnomalyDetector
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		n:         0,
		mean:      0,
		m2:        0,
		stdDev:    0,
		anomalies: []pb.FrequencyResponse{},
	}
}

// AddData adds new data point and updates mean and stdDev
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

// CheckAnomaly checks if the given frequency is an anomaly
func (ad *AnomalyDetector) CheckAnomaly(frequency float64, k float64) bool {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	if math.Abs(frequency-ad.mean) > k*ad.stdDev {
		return true
	}
	return false
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTransmitterClient(conn)

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
				detector.anomalies = append(detector.anomalies, *resp)
			}
		}

		log.Printf("Processed value: %f, mean: %f, stdDev: %f", resp.Frequency, detector.mean, detector.stdDev)
	}
}
