package chat_v1

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/anton0701/chat-server/grpc/pkg"
)

var (
	_ pkg.Validator = (*CreateChatRequest)(nil)
	_ pkg.Validator = (*DeleteChatRequest)(nil)
	_ pkg.Validator = (*SendMessageRequest)(nil)
)

// Validate
//
// Возвращает:
//   - error, если User_IDs пустой.
//   - error, если ChatName пустой либо состоит только из пробелов.
//   - nil в остальных случаях.
func (req *CreateChatRequest) Validate() error {
	// User_IDs должен содержать хотя бы 1 айди
	if len(req.User_IDs) == 0 {
		err := status.Error(codes.InvalidArgument, "User_IDs must contain at least one ID.")
		return err
	}

	// ChatName должен содеражть хотя бы 1 символ (не считая пробелов)
	trimmedChatName := strings.TrimSpace(req.ChatName)
	if len(trimmedChatName) == 0 {
		err := status.Error(codes.InvalidArgument, "Chat name must not be empty")
		return err
	}

	return nil
}

// Validate
//
// Возвращает:
//   - error, если ID чата не указан.
//   - nil в остальных случаях.
func (req *DeleteChatRequest) Validate() error {
	// В запросе должен содержаться ID чата
	if req.ID == 0 {
		err := status.Error(codes.InvalidArgument, "ID required")
		return err
	}

	return nil
}

// Validate
//
// Возвращает:
//   - error, если Text пустой или состоит только из пробелов.
//   - nil в остальных случаях.
func (req *SendMessageRequest) Validate() error {
	// MessageText должен содержать хотя бы 1 символ (не считая пробелов)
	trimmedMessage := strings.TrimSpace(req.Text)
	if len(trimmedMessage) == 0 {
		err := status.Error(codes.InvalidArgument, "Message text must contain at least 1 non-space character")
		return err
	}

	return nil
}
