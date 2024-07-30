package chat

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"

	desc "github.com/anton0701/chat-server/grpc/pkg/chat_v1"
	"github.com/anton0701/chat-server/internal/repository"
)

const (
	chatListTableName         = "chats"
	chatListIDColumn          = "id"
	chatListNameColumn        = "name"
	chatListDescriptionColumn = "description"

	chatUsersTableName    = "chat_users"
	chatUsersChatIDColumn = "chat_id"
	chatUsersUserIDColumn = "user_id"

	chatMessagesTableName       = "chat_messages"
	chatMessagesMessageIDColumn = "id"
	chatMessagesChatIDColumn    = "chat_id"
	chatMessagesUserIDColumn    = "user_id"
	chatMessagesMessageColumn   = "message"
	chatMessagesCreatedAtColumn = "created_at"
)

// TODO: видимо переделать на 3 репозитория, каждый работает со своей таблицей: chats, chat_users, chat_messages
type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) repository.ChatRepository {
	return &repo{db: db}
}

func (r *repo) CreateChat(ctx context.Context, req *desc.CreateChatRequest) (ID int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (r *repo) DeleteChat(ctx context.Context, req *desc.DeleteChatRequest) (err error) {
	//TODO implement me
	panic("implement me")
}

func (r *repo) AddMessage(ctx context.Context, req *desc.SendMessageRequest) (messageID int64, err error) {
	insertMessageBuilder := sq.
		Insert(chatMessagesTableName).
		PlaceholderFormat(sq.Dollar).
		Columns(chatMessagesChatIDColumn, chatMessagesUserIDColumn, chatMessagesMessageColumn, chatMessagesCreatedAtColumn).
		Values(req.Chat_ID, req.User_IDFrom, req.Text, req.Timestamp.AsTime()).
		Suffix("RETURNING id")

	query, args, err := insertMessageBuilder.ToSql()
	if err != nil {
		// TODO: откуда брать логгер? прокидывать его из сервиса или на месте создавать?
		//s.log.Error("Method Send-Message. Unable to create query to send message", zap.Error(err))
		return messageID, err
	}

	err = r.db.
		QueryRow(ctx, query, args...).
		Scan(&messageID)
	if err != nil {
		// TODO: откуда брать логгер? прокидывать его из сервиса или на месте создавать?
		//s.log.Error("Method Send-Message. Unable to execute query to send message", zap.Error(err))
		//return messageID, status.Errorf(codes.Internal, "Method Send-Message. Unable to execute query to send message, error: %v", err)
		return messageID, err
	}

	return messageID, err
}
