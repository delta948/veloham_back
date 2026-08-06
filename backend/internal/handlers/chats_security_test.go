package handlers

import (
	"net/http"
	"testing"

	"veloham/backend/internal/services"
)

func TestChatWebSocketOrigin(t *testing.T) {
	handler := NewChatHandler(nil, nil, services.JWTService{}, "https://veloham.example")
	for _, tt := range []struct {
		origin string
		want   bool
	}{
		{origin: "https://veloham.example", want: true},
		{origin: "https://evil.example", want: false},
		{origin: "", want: true},
	} {
		req, _ := http.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Origin", tt.origin)
		if got := handler.checkOrigin(req); got != tt.want {
			t.Fatalf("checkOrigin(%q) = %v, want %v", tt.origin, got, tt.want)
		}
	}
}

func TestWebSocketTokenUsesProtocolOrAuthorizationHeader(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Sec-WebSocket-Protocol", "access_token, token-from-protocol")
	if got := websocketToken(req); got != "token-from-protocol" {
		t.Fatalf("websocketToken(protocol) = %q", got)
	}

	req, _ = http.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Authorization", "Bearer token-from-header")
	if got := websocketToken(req); got != "token-from-header" {
		t.Fatalf("websocketToken(header) = %q", got)
	}

	req, _ = http.NewRequest(http.MethodGet, "/ws?token=leaked-token", nil)
	if got := websocketToken(req); got != "" {
		t.Fatalf("websocketToken() accepted query token %q", got)
	}
}
