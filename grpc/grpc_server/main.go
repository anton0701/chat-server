package main

import (
	"context"
	"fmt"
	desc "github.com/anton0701/chat-server/grpc/pkg/chat_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
)

const grpcPort = 50052

type server struct {
	desc.UnimplementedChatV1Server
}

func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	fmt.Printf("Method Create. Input params:\nUsernames: %s\n************\n\n", req.Usernames)
	resp := &desc.CreateResponse{
		Id: 1,
	}
	return resp, nil
}
func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	fmt.Printf("Method Delete. Input params:\nId: %d\n************\n\n", req.Id)
	return &emptypb.Empty{}, nil
}
func (s *server) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	fmt.Printf("Method SendMessage. Input params:\nFrom: %s\nText: %s\nTimestamp: %s************\n\n", req.GetFrom(), req.GetText(), req.GetTimestamp())
	return &emptypb.Empty{}, nil
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
