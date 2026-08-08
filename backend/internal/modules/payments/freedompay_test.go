package payments

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestSignatureIsStableAndCallbackIsVerified(t *testing.T) {
	params := map[string]string{"pg_amount": "30", "pg_order_id": "order-1", "pg_salt": "salt"}
	if got := sign("result", params, "secret"); got != "583932b5293db3579f588b4468861bdd" {
		t.Fatalf("unexpected signature: %s", got)
	}
	values := url.Values{"pg_amount": {"30"}, "pg_order_id": {"order-1"}, "pg_salt": {"salt"}}
	values.Set("pg_sig", sign("result", params, "secret"))
	client := FreedomPayClient{Secret: "secret"}
	if !client.VerifyCallback(values) {
		t.Fatal("valid callback signature rejected")
	}
	values.Set("pg_amount", "31")
	if client.VerifyCallback(values) {
		t.Fatal("tampered callback signature accepted")
	}
}

func TestInitAndStatus(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		script := r.URL.Path[1:]
		params := map[string]string{}
		for key, value := range r.Form {
			if key != "pg_sig" {
				params[key] = value[0]
			}
		}
		if r.Form.Get("pg_sig") != sign(script, params, "secret") {
			t.Errorf("invalid request signature for %s", script)
		}
		var body bytes.Buffer
		if script == "init_payment.php" {
			_ = xml.NewEncoder(&body).Encode(freedomPayResponse{Status: "ok", PaymentID: "provider-1", RedirectURL: "https://pay.example/1"})
		} else {
			_ = xml.NewEncoder(&body).Encode(freedomPayResponse{Status: "ok", PaymentID: "provider-1", PaymentStatus: "success", PaymentMethod: "bankcard", Captured: "1", Amount: "30.00", Currency: "KGS"})
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/xml"}}, Body: io.NopCloser(strings.NewReader(body.String()))}, nil
	})
	client := FreedomPayClient{BaseURL: "https://api.example", MerchantID: "merchant", Secret: "secret", HTTPClient: &http.Client{Transport: transport}}
	paymentID, redirect, err := client.Init(context.Background(), "order-1", "user-1", "u@example.test", "https://v.test/result", "https://v.test/payment")
	if err != nil || paymentID != "provider-1" || redirect != "https://pay.example/1" {
		t.Fatalf("Init() = %s, %s, %v", paymentID, redirect, err)
	}
	paid, err := client.Status(context.Background(), paymentID)
	if err != nil || !paid {
		t.Fatalf("Status() = %v, %v", paid, err)
	}
}
