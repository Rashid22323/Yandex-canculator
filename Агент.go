//Агент
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "path/to/calculator"
)

type server struct{}

func (s *server) Calculate(ctx context.Context, req *pb.Expression) (*pb.Result, error) {
	switch req.Operation {
	case "+":
		return &pb.Result{Value: req.Operand1 + req.Operand2}, nil
	case "-":
		return &pb.Result{Value: req.Operand1 - req.Operand2}, nil
	case "*":
		return &pb.Result{Value: req.Operand1 * req.Operand2}, nil
	case "/":
		if req.Operand2 == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return &pb.Result{Value: req.Operand1 / req.Operand2}, nil
	}
	return nil, fmt.Errorf("invalid operation")
}

func main() {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCalculatorServer(s, &server{})
	fmt.Println("Agent is running on port 8081")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

