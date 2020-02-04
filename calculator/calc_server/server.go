package main

import (
	"assignments/calculator/calcpb"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

type server struct{}

func (*server) AddNumbers(ctx context.Context, req *calcpb.AddRequest) (*calcpb.AddResponse, error) {
	fmt.Printf("Invoked adding of numbers %v and %v", req.GetX(), req.GetY())
	x := req.GetX()
	y := req.GetY()
	result := x + y
	res := &calcpb.AddResponse{
		Result: result,
	}

	return res, nil
}

func (*server) PrimeNumberDecomposition(req *calcpb.PrimeRequest, stream calcpb.CalculatorService_PrimeNumberDecompositionServer) error {
	fmt.Printf("PrimeNumberDecomposition was invoked with: %v\n", req.GetX())
	n := req.GetX()
	k := int64(2)
	for n > 1 {
		if n%k == 0 {
			stream.Send(&calcpb.PrimeResponse{PrimeFactor: int64(k)})
			time.Sleep(1000 * time.Millisecond)
			n = n / k
		} else {
			k = k + 1
		}
	}
	return nil
}

func (*server) ComputeAverage(stream calcpb.CalculatorService_ComputeAverageServer) error {
	fmt.Printf("Compute Average stream was invoked \n")
	numbers := []float32{}

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			// end of stream
			total := float32(0)
			for _, n := range numbers {
				total += n
			}
			return stream.SendAndClose(&calcpb.AverageResponse{
				Average: total / float32(len(numbers)),
			})
		}

		if err != nil {
			log.Fatalf("Failed to stream client data: %v", err)
		}

		numbers = append(numbers, msg.GetX())

	}
}

func (*server) FindMaximum(stream calcpb.CalculatorService_FindMaximumServer) error {
	fmt.Printf("Find Maximum bidi stream was invoked \n")
	max := int64(0)

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Failed to read client stream: %v", err)
		}

		n := req.GetX()

		if n > max {
			max = n
		}

		sendErr := stream.Send(&calcpb.MaximumResponse{
			Maximum: max,
		})

		if sendErr != nil {
			log.Fatalf("Error while responding to client stream: %v", err)
		}

	}
}

func main() {
	fmt.Println("Starting grpc server")

	lis, err := net.Listen("tcp", "localhost:50051")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calcpb.RegisterCalculatorServiceServer(s, &server{})

	s.Serve(lis)
}
