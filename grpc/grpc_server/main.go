package main

import (
	"context"
	"flag"
	"log"
	"net"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	config "github.com/anton0701/chat-server/config"
	env "github.com/anton0701/chat-server/config/env"
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

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}

func main() {
	flag.Parse()
	ctx := context.Background()
	// Инициализируем логгер
	logger, err := initLogger()
	if err != nil {
		log.Fatalf("%s\nUnable to init logger, error: %#v", grpcChatAPIDesc, err)
	}
	// Считываем переменные окружения (env vars)
	err = config.Load(configPath)
	if err != nil {
		logger.Fatal("Unable to init config", zap.Error(err))
	}

	grpcConfig, err := env.NewGRPCConfig()
	if err != nil {
		logger.Fatal("Unable to get grpc config", zap.Error(err))
	}

	pgConfig, err := env.NewPGConfig()
	if err != nil {
		logger.Fatal("Unable to get postgres config", zap.Error(err))
	}

	lis, err := net.Listen("tcp", grpcConfig.Address())
	if err != nil {
		logger.Panic("Failed to listen", zap.Error(err))
	}

	pool, err := pgxpool.Connect(ctx, pgConfig.DSN())
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
	var config = zap.NewProductionConfig()
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

	// TODO: текст ошибки вынести в константу?
	// Проверка необходимых полей в запросе, User_IDs должен содеражть хотя бы 1 айди
	if len(req.User_IDs) == 0 {
		err := status.Error(codes.InvalidArgument, "User_IDs must contain at least one ID.")
		s.log.Error("Method Create-Chat.", zap.Error(err))
		return nil, err
	}

	// Проверка необходимых полей в запросе, ChatName должен содеражть хотя бы 1 символ (не считая пробелов)
	trimmedChatName := strings.TrimSpace(req.ChatName)
	if len(trimmedChatName) == 0 {
		err := status.Error(codes.InvalidArgument, "Chat name must not be empty")
		s.log.Error("Method Create-Chat.", zap.Error(err))
		return nil, err
	}

	// Создаем транзакцию, чтобы выполнились все запросы к БД ИЛИ не выполнился ни один
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("Unable to start transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to start transaction: %v", err)
	}
	// Откатываем транзакцию в случае возникновения ошибки
	defer tx.Rollback(ctx)

	// Билдер первого INSERT запроса в таблицу "chats". Участвует в транзакции
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

	for _, userID := range req.User_IDs {
		// Билдер второго INSERT запроса в таблицу "chats"
		// Берем User_IDs из запроса и вставляем в таблицу "chat_users"
		// Участвует в транзакции
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
	}

	// Коммит транзакции
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

	// Проверка необходимых полей. В запросе должен содержаться ID чата
	if req.ID == 0 {
		err := status.Error(codes.InvalidArgument, "ID required")
		s.log.Error("Method Delete-Chat.", zap.Error(err))
		return nil, err
	}

	// Создаем транзакцию, чтобы выполнились все запросы к БД ИЛИ не выполнился ни один
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("Method Delete-Chat. Unable to start transaction. Error info", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Delete-Chat. Unable to start transaction. Error info: %v", err)
	}
	// Откатываем транзакцию в случае возникновения ошибки
	defer tx.Rollback(ctx)

	// Билдер запроса удаления чата из списка чатов. Участвует в транзакции
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

	// Билдер запроса удаления участников чата из chat_users. Участвует в транзакции
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
func (s *server) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	s.log.Info("Method Send-Message", zap.Any("Input params", req))

	// TODO: добавить проверку на наличие в БД chat_id и user_id из запроса?

	// Проверка необходмых полей запроса. MessageText должен содеражть хотя бы 1 символ (не считая пробелов)
	trimmedMessage := strings.TrimSpace(req.Text)
	if len(trimmedMessage) == 0 {
		err := status.Error(codes.InvalidArgument, "Message text must contain at least 1 non-space character")
		s.log.Error("Method Send-Message.", zap.Error(err))
		return nil, err
	}

	var messageID int64
	insertMessageBuilder := sq.
		Insert("chat_messages").
		PlaceholderFormat(sq.Dollar).
		Columns("chat_id", "user_id", "message", "created_at").
		Values(req.Chat_ID, req.User_IDFrom, req.Text, req.Timestamp.AsTime()).
		Suffix("RETURNING id")

	query, args, err := insertMessageBuilder.ToSql()
	if err != nil {
		s.log.Error("Method Send-Message. Unable to create query to send message", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Send-Message. Unable to create query to send message, error: %v", err)
	}

	err = s.pool.
		QueryRow(ctx, query, args...).
		Scan(&messageID)
	if err != nil {
		s.log.Error("Method Send-Message. Unable to execute query to send message", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Method Send-Message. Unable to execute query to send message, error: %v", err)
	}

	return &emptypb.Empty{}, nil
}
