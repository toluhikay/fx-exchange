package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/toluhikay/fx-exchange/internal/fx"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	fxProvider fx.FXProvider
}

func NewWebSocketHandler(fxProvider fx.FXProvider) *WebSocketHandler {
	return &WebSocketHandler{fxProvider: fxProvider}
}

func (h *WebSocketHandler) HandleFXRates(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	updates := h.fxProvider.SubscribeRates()
	defer func() {
		// Channel is closed by StartRateUpdates
	}()

	for rates := range updates {
		data, err := json.Marshal(rates)
		if err != nil {
			log.Printf("Failed to marshal rates: %v", err)
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Failed to send rates: %v", err)
			return
		}
	}
}
