package main

import (
	"context"
	"fmt"
	"log"
	"net"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	desc "github.com/anton0701/chat-server/grpc/pkg/chat_v1"
)

const (
	grpcPort        = 50052
	grpcChatAPIDesc = "Chat-API-v1"
	dbDSN           = "host=localhost port=54322 dbname=chat user=chat-user password=chat-password"
)

type server struct {
	desc.UnimplementedChatV1Server
	pool *pgxpool.Pool
	log  *zap.Logger
}

func main() {
	ctx := context.Background()

	logger, err := initLogger()
	if err != nil {
		log.Fatalf("%s\nUnable to init logger, error: %#v", grpcChatAPIDesc, err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logger.Panic("Failed to listen", zap.Error(err))
	}

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		logger.Panic("Unable to connect to db", zap.Error(err))
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterChatV1Server(s, &server{
		pool: pool,
		log:  logger,
	})
	logger.Info("Server listening at", zap.Any("Address", lis.Addr()))

	if err = s.Serve(lis); err != nil {
		logger.Panic("Failed to serve", zap.Error(err))
	}
}

func initLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("API", grpcChatAPIDesc))
	return logger, nil
}

func (s *server) CreateChat(ctx context.Context, req *desc.CreateChatRequest) (*desc.CreateChatResponse, error) {
	s.log.Info("Method Create-Chat", zap.Any("input params", req))

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("Unable to start transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to start transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	builderChatInsert := sq.
		Insert("chats").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "description").
		Values(req.ChatName, req.ChatDescription.GetValue()).
		Suffix("RETURNING id")

	query, args, err := builderChatInsert.ToSql()
	if err != nil {
		s.log.Error("Method Create-Chat. Unable to create query from builder to insert chat info", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Create-Chat. Unable to create query from builder to insert chat info: %v", err)
	}

	var chatID int64
	err = tx.
		QueryRow(ctx, query, args...).
		Scan(&chatID)
	if err != nil {
		s.log.Error("Method Create-Chat. Unable to execute INSERT chat query", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Create-Chat. Unable to execute INSERT chat query, error: %v", err)
	}

	s.log.Info("Method Create-Chat. Insert chat",
		zap.Int64("chat_id", chatID))

	for _, userID := range req.User_IDs {
		builderChatUsersInsert := sq.
			Insert("chat_users").
			Columns("chat_id", "user_id").
			Values(chatID, userID).
			PlaceholderFormat(sq.Dollar)

		query, args, err := builderChatUsersInsert.ToSql()

		if err != nil {
			s.log.Error("Method Create-Chat. Unable to create query from builder to insert chat users", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "Method Create-Chat. Unable to create query from builder to insert chat users: %v", err)
		}

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			s.log.Error("Method Create-Chat. Unable to execute query from builder to insert chat users", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "Method Create-Chat. Unable to execute query from builder to insert chat users: %v", err)
		}

		s.log.Info("Method Create-Chat. Insert chat users",
			zap.Int64("chat_id", chatID),
			zap.Int64("user_id", userID))
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("Unable to commit transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to commit transaction: %v", err)
	}

	return &desc.CreateChatResponse{
		ID: chatID,
	}, nil
}

func (s *server) DeleteChat(ctx context.Context, req *desc.DeleteChatRequest) (*emptypb.Empty, error) {
	s.log.Info("Method Delete-Chat", zap.Any("Input params", req))

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("Method Delete-Chat. Unable to start transaction. Error info", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Delete-Chat. Unable to start transaction. Error info: %v", err)
	}
	defer tx.Rollback(ctx)

	deleteChatBuilder := sq.
		Delete("chats").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.ID})

	query, args, err := deleteChatBuilder.ToSql()
	if err != nil {
		s.log.Error("Method Delete-Chat. Unable to create query from builder to delete chat", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Delete-Chat. Unable to create query from builder to delete chat, error: %v", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		s.log.Error("Method Delete-Chat. Unable to execute query to delete chat", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Delete-Chat. Unable to execute query to delete chat, error: %v", err)
	}

	deleteChatUsersBuilder := sq.
		Delete("chat_users").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"chat_id": req.ID})
	query, args, err = deleteChatUsersBuilder.ToSql()
	if err != nil {
		s.log.Error("Method Delete-Chat. Unable to create query to delete chat users", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Delete-Chat. Unable to create query to delete chat users, error: %v", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		s.log.Error("Method Delete-Chat. Unable to execute query to delete chat users", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Delete-Chat. Unable to execute query to delete chat users, error: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("Method Delete-Chat. Unable to commit transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Delete-Chat. Unable to commit transaction, error: %v", err)
	}

	return &emptypb.Empty{}, nil
}
func (s *server) SendMessage(_ context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	log.Printf("%s\nMethod SendMessage.\nInput params:\n%+v\n************\n\n", grpcChatAPIDesc, req)

	return &emptypb.Empty{}, nil
}
