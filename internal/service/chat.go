package service

import (
	"context"
	"log/slog"

	"dooz/entity"
	"dooz/internal/app/api/dto"
	appErrors "dooz/internal/errors"
	wsHub "dooz/internal/infrastructure/websocket"
	boardRepo "dooz/internal/repository/board"
	chatRepo "dooz/internal/repository/chat_message"
)

type ChatService interface {
	SendDM(ctx context.Context, senderID, receiverID, content string) (*dto.ChatMessageDTO, error)
	SendGameChat(ctx context.Context, senderID, boardID, content string) (*dto.ChatMessageDTO, error)
	GetDMHistory(ctx context.Context, userA, userB string, limit int, before string) ([]*dto.ChatMessageDTO, error)
	GetGameChatHistory(ctx context.Context, boardID string) ([]*dto.ChatMessageDTO, error)
}

type chatService struct {
	chatRepo  chatRepo.Repository
	boardRepo boardRepo.Repository
	hub       *wsHub.Hub
	logger    *slog.Logger
}

func NewChatService(chatRepo chatRepo.Repository, boardRepo boardRepo.Repository, hub *wsHub.Hub, logger *slog.Logger) ChatService {
	return &chatService{
		chatRepo:  chatRepo,
		boardRepo: boardRepo,
		hub:       hub,
		logger:    logger.With("layer", "ChatService"),
	}
}

func (s *chatService) SendDM(ctx context.Context, senderID, receiverID, content string) (*dto.ChatMessageDTO, error) {
	receiver := receiverID
	msg := &entity.ChatMessage{
		SenderID:   senderID,
		ReceiverID: &receiver,
		Content:    content,
	}
	if err := s.chatRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	s.hub.SendToUsers([]string{receiverID}, wsHub.TypeChat, dto.ChatRealtimePayload{
		ChatType: "dm",
		Message:  msg.ToDTO(),
	})

	return msg.ToDTO(), nil
}

func (s *chatService) SendGameChat(ctx context.Context, senderID, boardID, content string) (*dto.ChatMessageDTO, error) {
	boardState, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if senderID != boardState.PlayerXID && senderID != boardState.PlayerOID {
		return nil, appErrors.ErrForbidden
	}

	board := boardID
	msg := &entity.ChatMessage{
		SenderID: senderID,
		BoardID:  &board,
		Content:  content,
	}
	if err := s.chatRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	recipients := []string{boardState.PlayerXID}
	if boardState.PlayerOID != "" && boardState.PlayerOID != boardState.PlayerXID {
		recipients = append(recipients, boardState.PlayerOID)
	}
	s.hub.SendToUsers(recipients, wsHub.TypeChat, dto.ChatRealtimePayload{
		ChatType: "game",
		BoardID:  boardID,
		Message:  msg.ToDTO(),
	})

	return msg.ToDTO(), nil
}

func (s *chatService) GetDMHistory(ctx context.Context, userA, userB string, limit int, before string) ([]*dto.ChatMessageDTO, error) {
	messages, err := s.chatRepo.GetDMHistory(ctx, userA, userB, limit, before)
	if err != nil {
		return nil, err
	}
	dtos := make([]*dto.ChatMessageDTO, len(messages))
	for i, m := range messages {
		dtos[i] = m.ToDTO()
	}
	return dtos, nil
}

func (s *chatService) GetGameChatHistory(ctx context.Context, boardID string) ([]*dto.ChatMessageDTO, error) {
	messages, err := s.chatRepo.GetGameChatHistory(ctx, boardID)
	if err != nil {
		return nil, err
	}
	dtos := make([]*dto.ChatMessageDTO, len(messages))
	for i, m := range messages {
		dtos[i] = m.ToDTO()
	}
	return dtos, nil
}
