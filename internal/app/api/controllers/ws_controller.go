package controllers

import (
	"log/slog"
	"net/http"

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

	go client.WritePump()

	// ReadPump: keep connection alive, handle incoming pings
	defer func() {
		c.hub.Unregister(client)
		conn.Close()
		_ = c.userService.SetOnline(ctx.Request.Context(), userIDStr, false)
		c.logger.Info("WebSocket client disconnected", "userID", userIDStr)
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
