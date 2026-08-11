package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/websocket"

	orderv1 "ride-hailing/services/order-service/api/order/v1"
)

type orderChatMessage struct {
	Type        string `json:"type"`
	OrderID     string `json:"order_id"`
	SenderRole  string `json:"sender_role"`
	SenderID    string `json:"sender_id"`
	Content     string `json:"content"`
	ClientMsgID string `json:"client_msg_id,omitempty"`
	SentAt      string `json:"sent_at"`
}

type orderChatClient struct {
	orderID string
	role    string
	userID  int64
	conn    *websocket.Conn
	send    chan orderChatMessage
	done    chan struct{}
}

type orderChatHub struct {
	mu    sync.RWMutex
	rooms map[string]map[*orderChatClient]struct{}
}

type orderChatOrderService interface {
	GetOrderDetail(context.Context, int64, int64) (*orderv1.GetOrderDetailReply, error)
}

func newOrderChatHub() *orderChatHub {
	return &orderChatHub{rooms: make(map[string]map[*orderChatClient]struct{})}
}

func (h *orderChatHub) register(client *orderChatClient) {
	if h == nil || client == nil || strings.TrimSpace(client.orderID) == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[client.orderID] == nil {
		h.rooms[client.orderID] = make(map[*orderChatClient]struct{})
	}
	h.rooms[client.orderID][client] = struct{}{}
}

func (h *orderChatHub) unregister(client *orderChatClient) {
	if h == nil || client == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.rooms[client.orderID]
	if clients == nil {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(h.rooms, client.orderID)
	}
}

func (h *orderChatHub) broadcast(orderID string, msg orderChatMessage) int {
	if h == nil || strings.TrimSpace(orderID) == "" {
		return 0
	}
	h.mu.RLock()
	clients := make([]*orderChatClient, 0, len(h.rooms[orderID]))
	for client := range h.rooms[orderID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	delivered := 0
	for _, client := range clients {
		select {
		case client.send <- msg:
			delivered++
		default:
		}
	}
	return delivered
}

func registerOrderChatRoutes(srv *khttp.Server, hub *orderChatHub, orderSvc orderChatOrderService) {
	if hub == nil {
		hub = newOrderChatHub()
	}
	router := srv.Route("/")
	for _, path := range []string{
		"/carpool/orders/{id}/chat/ws",
		"/api/v1/driver/orders/{id}/chat/ws",
		"/api/v1/passenger/orders/{id}/chat/ws",
	} {
		route := path
		router.GET(route, func(ctx khttp.Context) error {
			return handleOrderChatWebSocket(ctx, hub, orderSvc)
		})
	}
}

func handleOrderChatWebSocket(ctx khttp.Context, hub *orderChatHub, orderSvc orderChatOrderService) error {
	applyOrderChatQueryToken(ctx.Request())
	orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
	if err != nil {
		return returnBadRequest(ctx, invalidOrderIDMessage)
	}
	role := normalizeOrderChatRole(ctx.Query().Get("role"))
	if role == "" {
		return returnBadRequest(ctx, "role must be driver or passenger")
	}
	userID := currentUserID(ctx.Request())
	if userID <= 0 {
		userID = int64(parseInt(ctx.Query().Get("user_id")))
	}
	if userID <= 0 {
		return returnBadRequest(ctx, "user_id is required")
	}
	if orderSvc != nil {
		if _, err := orderSvc.GetOrderDetail(ctx, orderID, userID); err != nil {
			return err
		}
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return err
	}
	client := &orderChatClient{
		orderID: int64String(orderID),
		role:    role,
		userID:  userID,
		conn:    conn,
		send:    make(chan orderChatMessage, 16),
		done:    make(chan struct{}),
	}
	hub.register(client)
	defer func() {
		hub.unregister(client)
		close(client.done)
		_ = conn.Close()
	}()

	done := make(chan struct{})
	go writeOrderChatMessages(client, done)
	readOrderChatMessages(client, hub)
	<-done
	return nil
}

func readOrderChatMessages(client *orderChatClient, hub *orderChatHub) {
	for {
		var incoming orderChatMessage
		if err := client.conn.ReadJSON(&incoming); err != nil {
			return
		}
		content := strings.TrimSpace(incoming.Content)
		if incoming.Type != "" && incoming.Type != "chat" {
			continue
		}
		if content == "" {
			continue
		}
		if len([]rune(content)) > 1000 {
			content = string([]rune(content)[:1000])
		}
		hub.broadcast(client.orderID, orderChatMessage{
			Type:        "chat",
			OrderID:     client.orderID,
			SenderRole:  client.role,
			SenderID:    int64String(client.userID),
			Content:     content,
			ClientMsgID: strings.TrimSpace(incoming.ClientMsgID),
			SentAt:      time.Now().Format(time.RFC3339),
		})
	}
}

func writeOrderChatMessages(client *orderChatClient, done chan<- struct{}) {
	defer close(done)
	for {
		select {
		case <-client.done:
			return
		case msg := <-client.send:
			if err := client.conn.WriteJSON(msg); err != nil {
				return
			}
		}
	}
}

func normalizeOrderChatRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "driver", "passenger":
		return strings.ToLower(strings.TrimSpace(role))
	default:
		return ""
	}
}

func applyOrderChatQueryToken(r *http.Request) {
	if r == nil || strings.TrimSpace(r.Header.Get("Authorization")) != "" {
		return
	}
	token := strings.TrimSpace(r.URL.Query().Get("access_token"))
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if token == "" {
		return
	}
	if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = "Bearer " + token
	}
	r.Header.Set("Authorization", token)
}
