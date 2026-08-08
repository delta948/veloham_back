package payments

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type FreedomPayClient struct {
	BaseURL, MerchantID, Secret string
	HTTPClient                  *http.Client
}

type freedomPayResponse struct {
	Status           string `xml:"pg_status"`
	PaymentID        string `xml:"pg_payment_id"`
	RedirectURL      string `xml:"pg_redirect_url"`
	PaymentStatus    string `xml:"pg_payment_status"`
	PaymentMethod    string `xml:"pg_payment_method"`
	Captured         string `xml:"pg_captured"`
	Amount           string `xml:"pg_amount"`
	Currency         string `xml:"pg_currency"`
	Salt             string `xml:"pg_salt"`
	Signature        string `xml:"pg_sig"`
	ErrorDescription string `xml:"pg_error_description"`
}

func (c FreedomPayClient) Init(ctx context.Context, orderID, userID, email, resultURL, returnURL string) (string, string, error) {
	params := map[string]string{
		"pg_order_id": orderID, "pg_merchant_id": c.MerchantID, "pg_amount": "30",
		"pg_currency": "KGS", "pg_description": "Размещение объявления VELOHAM",
		"pg_salt": randomSalt(), "pg_result_url": resultURL,
		"pg_success_url": returnURL, "pg_failure_url": returnURL,
		"pg_request_method": "POST", "pg_success_url_method": "GET", "pg_failure_url_method": "GET",
		"pg_user_id": userID, "pg_language": "ru",
	}
	if email != "" {
		params["pg_user_contact_email"] = email
	}
	params["pg_sig"] = sign("init_payment.php", params, c.Secret)
	var response freedomPayResponse
	if err := c.post(ctx, "/init_payment.php", params, &response); err != nil {
		return "", "", err
	}
	if response.Status != "ok" || response.PaymentID == "" || response.RedirectURL == "" {
		return "", "", fmt.Errorf("freedompay init failed: %s", response.ErrorDescription)
	}
	return response.PaymentID, response.RedirectURL, nil
}

func (c FreedomPayClient) Status(ctx context.Context, paymentID string) (bool, error) {
	params := map[string]string{"pg_merchant_id": c.MerchantID, "pg_payment_id": paymentID, "pg_salt": randomSalt()}
	params["pg_sig"] = sign("get_status3.php", params, c.Secret)
	var response freedomPayResponse
	if err := c.post(ctx, "/get_status3.php", params, &response); err != nil {
		return false, err
	}
	if response.Status != "ok" {
		return false, fmt.Errorf("freedompay status failed: %s", response.ErrorDescription)
	}
	if response.PaymentID != paymentID || !validListingFee(response.Amount) || response.Currency != "KGS" {
		return false, errors.New("freedompay returned mismatched payment details")
	}
	paid := response.PaymentStatus == "success" || response.PaymentStatus == "ok"
	if response.PaymentMethod == "bankcard" && response.Captured == "0" {
		paid = false
	}
	return paid, nil
}

func validListingFee(value string) bool {
	amount, err := strconv.ParseFloat(value, 64)
	return err == nil && amount == ListingPlacementPrice
}

func (c FreedomPayClient) VerifyCallback(values url.Values) bool {
	provided := values.Get("pg_sig")
	params := make(map[string]string, len(values))
	for key, value := range values {
		if key != "pg_sig" && len(value) > 0 {
			params[key] = value[0]
		}
	}
	expected := sign("result", params, c.Secret)
	if len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func (c FreedomPayClient) post(ctx context.Context, path string, params map[string]string, target any) error {
	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("freedompay HTTP status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if err := xml.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode freedompay response: %w", err)
	}
	return nil
}

func sign(script string, params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key != "pg_sig" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	parts := []string{script}
	for _, key := range keys {
		parts = append(parts, params[key])
	}
	parts = append(parts, secret)
	sum := md5.Sum([]byte(strings.Join(parts, ";")))
	return hex.EncodeToString(sum[:])
}

func randomSalt() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(value)
}
