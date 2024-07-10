package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	desc "github.com/anton0701/chat-server/grpc/pkg/chat_v1"
)

const (
	grpcPort        = 50052
	grpcChatApiDesc = "Chat-Api-v1"
)

type server struct {
	desc.UnimplementedChatV1Server
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterChatV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) Create(_ context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	log.Printf("%s\nMethod Create.\nInput params:\n%+v\n************\n\n", grpcChatApiDesc, req)

	return &desc.CreateResponse{
		Id: 1,
	}, nil
}

func (s *server) Delete(_ context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	log.Printf("%s\nMethod Delete.\nInput params:\n%+v\n************\n\n", grpcChatApiDesc, req)

	return &emptypb.Empty{}, nil
}

func (s *server) SendMessage(_ context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	log.Printf("%s\nMethod SendMessage.\nInput params:\n%+v\n************\n\n", grpcChatApiDesc, req)

	return &emptypb.Empty{}, nil
}
