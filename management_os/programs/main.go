package main

import (
	context "context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "programs/api/servicea"
)

type grpcResponder struct {
	pb.UnimplementedGreeterServer
}

func (g grpcResponder) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	panic("implement me")
}

func (g grpcResponder) SayHelloAgain(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	panic("implement me")
}

func main() {

	log.Printf("Hello from go!\n")

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:50051"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGreeterServer(grpcServer, &grpcResponder{})

	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}
