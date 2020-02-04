package main

import (
	"assignments/calculator/calcpb"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		fmt.Println("Couldn't connect", err)
	}
	defer conn.Close()

	client := calcpb.NewCalculatorServiceClient(conn)

	//doUnary(client)
	//doServerStreaming(client)
	//doClientStreaming(client)
	doBiDiStreaming(client)
}

func doUnary(c calcpb.CalculatorServiceClient) {
	res, _ := c.AddNumbers(context.Background(), &calcpb.AddRequest{X: 3, Y: 10})
	fmt.Printf("Response from Add: %v", res.Result)
}

func doServerStreaming(c calcpb.CalculatorServiceClient) {
	fmt.Println("Starting Server Streaming RPC")
	resStream, err := c.PrimeNumberDecomposition(context.Background(), &calcpb.PrimeRequest{X: 120})

	if err != nil {
		log.Fatalf("Fatal error on stream init: %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			// end of stream, jump out.
			fmt.Println("Reached end of stream")
			break
		}

		if err != nil {
			log.Fatalf("Fatal error reading stream: %v", err)
		}

		fmt.Printf("Result from stream: %v \n", msg.GetPrimeFactor())
	}
}

func doClientStreaming(c calcpb.CalculatorServiceClient) {
	fmt.Println("Starting Client Streaming RPC")

	requests := []*calcpb.AverageRequest{
		&calcpb.AverageRequest{
			X: 1,
		},
		&calcpb.AverageRequest{
			X: 2,
		},
		&calcpb.AverageRequest{
			X: 3,
		},
		&calcpb.AverageRequest{
			X: 4,
		},
	}

	stream, err := c.ComputeAverage(context.Background())

	if err != nil {
		log.Fatalf("Fatal error while calling avarage calc: %v", err)
	}

	// send sliced requests
	for _, req := range requests {
		fmt.Printf("Sending req: %v\n", req)
		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("Fatal error while getting response from avg calc: %v", err)
	}
	fmt.Printf("Response from avg calc: %v", res)

}

func doBiDiStreaming(c calcpb.CalculatorServiceClient) {
	fmt.Println("Starting BiDi Streaming RPC")

	requests := []*calcpb.MaximumRequest{
		&calcpb.MaximumRequest{
			X: 1,
		},
		&calcpb.MaximumRequest{
			X: 10,
		},
		&calcpb.MaximumRequest{
			X: 5,
		},
		&calcpb.MaximumRequest{
			X: 13,
		},
		&calcpb.MaximumRequest{
			X: 4,
		},
	}

	// create stream
	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Fatal error while setting stream: %v", err)
		return
	}

	waitc := make(chan struct{})

	// send messages
	go func() {
		// func to send messages
		for _, req := range requests {
			fmt.Printf("Sending message: %v", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	// receive messages
	go func() {
		// func to receive messages

		for {
			res, err := stream.Recv()
			// end of stream, close.
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalf("Fatal error while receiving stream: %v", err)
				break
			}

			fmt.Printf("Current maximum is: %v \n", res)
		}
		close(waitc)
	}()

	// block until finished
	<-waitc
}
