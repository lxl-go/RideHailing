package server

import (
	"net/http/httptest"
	"testing"
)

func TestOrderChatHubBroadcastsWithinOrderRoom(t *testing.T) {
	hub := newOrderChatHub()
	driver := &orderChatClient{orderID: "1001", role: "driver", userID: 3, send: make(chan orderChatMessage, 1)}
	passenger := &orderChatClient{orderID: "1001", role: "passenger", userID: 6, send: make(chan orderChatMessage, 1)}
	otherOrder := &orderChatClient{orderID: "2002", role: "passenger", userID: 8, send: make(chan orderChatMessage, 1)}

	hub.register(driver)
	hub.register(passenger)
	hub.register(otherOrder)

	msg := orderChatMessage{Type: "chat", OrderID: "1001", SenderRole: "driver", SenderID: "3", Content: "arriving soon"}
	delivered := hub.broadcast("1001", msg)

	if delivered != 2 {
		t.Fatalf("broadcast delivered %d clients, want 2", delivered)
	}
	assertChatMessage(t, driver.send, msg)
	assertChatMessage(t, passenger.send, msg)

	select {
	case got := <-otherOrder.send:
		t.Fatalf("other order received message: %#v", got)
	default:
	}
}

func TestOrderChatHubUnregisterRemovesClient(t *testing.T) {
	hub := newOrderChatHub()
	driver := &orderChatClient{orderID: "1001", role: "driver", userID: 3, send: make(chan orderChatMessage, 1)}
	passenger := &orderChatClient{orderID: "1001", role: "passenger", userID: 6, send: make(chan orderChatMessage, 1)}

	hub.register(driver)
	hub.register(passenger)
	hub.unregister(driver)

	msg := orderChatMessage{Type: "chat", OrderID: "1001", SenderRole: "passenger", SenderID: "6", Content: "ok"}
	delivered := hub.broadcast("1001", msg)

	if delivered != 1 {
		t.Fatalf("broadcast delivered %d clients, want 1", delivered)
	}
	assertChatMessage(t, passenger.send, msg)

	select {
	case got := <-driver.send:
		t.Fatalf("unregistered client received message: %#v", got)
	default:
	}
}

func TestApplyOrderChatQueryTokenSetsAuthorizationHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/passenger/orders/1001/chat/ws?access_token=abc123", nil)

	applyOrderChatQueryToken(req)

	if got := req.Header.Get("Authorization"); got != "Bearer abc123" {
		t.Fatalf("Authorization = %q, want %q", got, "Bearer abc123")
	}
}

func TestApplyOrderChatQueryTokenKeepsExistingAuthorizationHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/passenger/orders/1001/chat/ws?access_token=abc123", nil)
	req.Header.Set("Authorization", "Bearer existing")

	applyOrderChatQueryToken(req)

	if got := req.Header.Get("Authorization"); got != "Bearer existing" {
		t.Fatalf("Authorization = %q, want %q", got, "Bearer existing")
	}
}

func assertChatMessage(t *testing.T, ch <-chan orderChatMessage, want orderChatMessage) {
	t.Helper()
	select {
	case got := <-ch:
		if got != want {
			t.Fatalf("message = %#v, want %#v", got, want)
		}
	default:
		t.Fatalf("no chat message received")
	}
}
