package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	pb "github.com/HSczy/gRPCLearning/route"
)

func main{
	lis, err := net.Listen("tcp", "localhost:5000")
	if err != nil {
		log.Fatalln("cannot create a listener at the address")
	}
	grpcServer := grpc.NewServer()
	pb.

}