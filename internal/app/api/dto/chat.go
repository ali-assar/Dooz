package dto

type ChatMessageDTO struct {
	ID         string `json:"id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id,omitempty"`
	BoardID    string `json:"board_id,omitempty"`
	Content    string `json:"content"`
	ReadAt     int64  `json:"read_at,omitempty"`
	CreatedAt  int64  `json:"created_at"`
}

type SendDMRequest struct {
	ReceiverID string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required,min=1,max=1000"`
}

type SendGameChatRequest struct {
	BoardID string `json:"board_id" binding:"required"`
	Content string `json:"content" binding:"required,min=1,max=500"`
}

type ChatHistoryRequest struct {
	UserID string `uri:"user_id" binding:"required"`
	Limit  int    `form:"limit,default=50" binding:"omitempty,min=1,max=100"`
	Before int64  `form:"before"`
}

type GameChatHistoryRequest struct {
	BoardID string `uri:"board_id" binding:"required"`
}
