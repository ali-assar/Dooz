package service

import (
	"context"
	"log/slog"
	"time"

	"dooz/entity"
	"dooz/internal/app/api/dto"
	wsHub "dooz/internal/infrastructure/websocket"
	chatRepo "dooz/internal/repository/chat_message"
)

type ChatService interface {
	SendDM(ctx context.Context, senderID, receiverID, content string) (*dto.ChatMessageDTO, error)
	SendGameChat(ctx context.Context, senderID, boardID, content string) (*dto.ChatMessageDTO, error)
	GetDMHistory(ctx context.Context, userA, userB string, limit int, before int64) ([]*dto.ChatMessageDTO, error)
	GetGameChatHistory(ctx context.Context, boardID string) ([]*dto.ChatMessageDTO, error)
}

type chatService struct {
	chatRepo chatRepo.Repository
	hub      *wsHub.Hub
	logger   *slog.Logger
}

func NewChatService(chatRepo chatRepo.Repository, hub *wsHub.Hub, logger *slog.Logger) ChatService {
	return &chatService{
		chatRepo: chatRepo,
		hub:      hub,
		logger:   logger.With("layer", "ChatService"),
	}
}

func (s *chatService) SendDM(ctx context.Context, senderID, receiverID, content string) (*dto.ChatMessageDTO, error) {
	msg := &entity.ChatMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		CreatedAt:  time.Now().Unix(),
	}
	if err := s.chatRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	s.hub.SendToUsers([]string{receiverID}, wsHub.TypeChat, msg.ToDTO())

	return msg.ToDTO(), nil
}

func (s *chatService) SendGameChat(ctx context.Context, senderID, boardID, content string) (*dto.ChatMessageDTO, error) {
	msg := &entity.ChatMessage{
		SenderID:  senderID,
		BoardID:   boardID,
		Content:   content,
		CreatedAt: time.Now().Unix(),
	}
	if err := s.chatRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	return msg.ToDTO(), nil
}

func (s *chatService) GetDMHistory(ctx context.Context, userA, userB string, limit int, before int64) ([]*dto.ChatMessageDTO, error) {
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
