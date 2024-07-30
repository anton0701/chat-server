package repository

import (
	"context"

	desc "github.com/anton0701/chat-server/grpc/pkg/chat_v1"
)

// TODO: видимо переделать на 3 репозитория, каждый работает со своей таблицей: chats, chat_users, chat_messages
type ChatRepository interface {
	CreateChat(ctx context.Context, req *desc.CreateChatRequest) (ID int64, err error)
	DeleteChat(ctx context.Context, req *desc.DeleteChatRequest) (err error)
	AddMessage(ctx context.Context, req *desc.SendMessageRequest) (messageID int64, err error)
}
