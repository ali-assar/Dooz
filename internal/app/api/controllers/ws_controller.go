package controllers

import (
	"log/slog"
	"net/http"
	"time"

	wsHub "dooz/internal/infrastructure/websocket"
	"dooz/internal/service"

	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
)

var upgrader = gorillaws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSController struct {
	hub         *wsHub.Hub
	userService service.UserService
	logger      *slog.Logger
}

func NewWSController(hub *wsHub.Hub, userService service.UserService, logger *slog.Logger) *WSController {
	return &WSController{
		hub:         hub,
		userService: userService,
		logger:      logger.With("layer", "WSController"),
	}
}

// HandleWS upgrades the connection to WebSocket and registers the client in the hub.
// WebSocket server emits:
// - `presence` payload: { user_id, is_online, last_seen_at }
// - `chat` payload (DM): { chat_type:"dm", message: ChatMessageDTO }
// - `chat` payload (game): { chat_type:"game", board_id, message: ChatMessageDTO }
//
//	@Summary	WebSocket connection
//	@Tags		ws
//	@Security	BearerAuth
//	@Router		/ws [get]
func (c *WSController) HandleWS(ctx *gin.Context) {
	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		c.logger.Error("failed to upgrade WebSocket", "error", err)
		return
	}

	client := &wsHub.Client{
		UserID: userIDStr,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	c.hub.Register(client)
	_ = c.userService.SetOnline(ctx.Request.Context(), userIDStr, true)
	c.broadcastPresence(userIDStr, true, 0)

	go client.WritePump()

	// ReadPump: keep connection alive, handle incoming pings
	defer func() {
		c.hub.Unregister(client)
		conn.Close()
		_ = c.userService.SetOnline(ctx.Request.Context(), userIDStr, false)
		c.broadcastPresence(userIDStr, false, time.Now().Unix())
		c.logger.Info("WebSocket client disconnected", "userID", userIDStr)
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *WSController) broadcastPresence(userID string, isOnline bool, lastSeenAt int64) {
	connected := c.hub.ConnectedUserIDs()
	if len(connected) == 0 {
		return
	}

	c.hub.SendToUsers(connected, wsHub.TypePresence, map[string]interface{}{
		"user_id":      userID,
		"is_online":    isOnline,
		"last_seen_at": lastSeenAt,
	})
}
