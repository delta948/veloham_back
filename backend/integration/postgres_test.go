package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"veloham/backend/internal/config"
	"veloham/backend/internal/models"
	"veloham/backend/internal/routes"
	"veloham/backend/internal/services"
	"veloham/backend/migrations"
)

func TestPostgresMigrationsAndAuthorization(t *testing.T) {
	db := newTestDatabase(t)

	t.Run("all migrations are recorded", func(t *testing.T) {
		var count int64
		if err := db.Table("schema_migrations").Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 19 {
			t.Fatalf("expected 19 applied migrations, got %d", count)
		}
	})

	owner := createUser(t, db, "integration-owner@veloham.test", "user")
	other := createUser(t, db, "integration-other@veloham.test", "user")
	admin := createUser(t, db, "integration-admin@veloham.test", "admin")
	listing := models.Listing{
		UserID: owner.ID, Title: "Integration listing", Description: "Authorization boundary",
		Price: 10000, InitialPrice: 10000, City: "Бишкек", Category: "MTB",
		Condition: "хорошее", DealType: "продажа", Status: "active", Labels: []string{},
	}
	if err := db.Create(&listing).Error; err != nil {
		t.Fatal(err)
	}

	t.Run("price history is immutable", func(t *testing.T) {
		row := models.ListingPriceHistory{
			ListingID: listing.ID, OldPrice: 10000, NewPrice: 9000,
			ChangedBy: owner.ID, ChangedAt: time.Now(),
		}
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Model(&row).Update("new_price", 8000).Error; err == nil {
			t.Fatal("expected immutable price history update to fail")
		}
	})

	jwt := services.NewJWTService("integration-test-secret-at-least-32-characters")
	ownerToken, _ := jwt.Generate(owner.ID)
	otherToken, _ := jwt.Generate(other.ID)
	adminToken, _ := jwt.Generate(admin.ID)
	quotaUser := createUser(t, db, "integration-quota@veloham.test", "user")
	quotaToken, _ := jwt.Generate(quotaUser.ID)
	router := routes.Setup(db, config.Config{
		Environment: "test", JWTSecret: "integration-test-secret-at-least-32-characters",
		CORSOrigin: "http://localhost:5173", UploadDir: t.TempDir(),
	})

	var ownerNotificationID, otherNotificationID string
	if err := db.Raw(`INSERT INTO notifications (user_id, type, message, link)
		VALUES (?, 'test', 'owner notification', '/') RETURNING id`, owner.ID).Scan(&ownerNotificationID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Raw(`INSERT INTO notifications (user_id, type, message, link)
		VALUES (?, 'test', 'other notification', '/') RETURNING id`, other.ID).Scan(&otherNotificationID).Error; err != nil {
		t.Fatal(err)
	}

	t.Run("notifications are isolated by user", func(t *testing.T) {
		response := request(router, http.MethodGet, "/api/v1/notifications", ownerToken)
		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
		}
		var rows []models.Notification
		if err := json.Unmarshal(response.Body.Bytes(), &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 || rows[0].UserID != owner.ID || strings.Contains(response.Body.String(), "other notification") {
			t.Fatalf("unexpected notifications payload: %s", response.Body.String())
		}

		response = request(router, http.MethodPatch, "/api/v1/notifications/"+otherNotificationID+"/read", ownerToken)
		if response.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for another user's notification, got %d", response.Code)
		}
		response = request(router, http.MethodPatch, "/api/v1/notifications/"+ownerNotificationID+"/read", ownerToken)
		if response.Code != http.StatusNoContent {
			t.Fatalf("expected 204 for own notification, got %d", response.Code)
		}
	})

	t.Run("listing owner boundary is enforced", func(t *testing.T) {
		response := request(router, http.MethodDelete, "/api/v1/listings/"+listing.ID, otherToken)
		if response.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", response.Code, response.Body.String())
		}
	})

	t.Run("user can upload an avatar", func(t *testing.T) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		file, err := writer.CreateFormFile("avatar", "avatar.png")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = file.Write([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"))
		_ = writer.Close()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", &body)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"avatar_url":"/uploads/avatar-`) {
			t.Fatalf("expected uploaded avatar, got %d: %s", response.Code, response.Body.String())
		}
	})

	t.Run("only three listing placements are free", func(t *testing.T) {
		for number := 1; number <= 4; number++ {
			response := createListingRequest(router, quotaToken, number)
			expected := http.StatusCreated
			if number == 4 {
				expected = http.StatusAccepted
			}
			if response.Code != expected {
				t.Fatalf("listing %d: expected %d, got %d: %s", number, expected, response.Code, response.Body.String())
			}
			if number == 4 {
				var payload struct {
					PaymentRequired bool   `json:"payment_required"`
					PaymentID       string `json:"payment_id"`
					Amount          int    `json:"amount"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
					t.Fatal(err)
				}
				if !payload.PaymentRequired || payload.PaymentID == "" || payload.Amount != 30 {
					t.Fatalf("unexpected payment response: %s", response.Body.String())
				}
			}
		}
		var free, pending int64
		db.Model(&models.ListingPlacement{}).Where("user_id = ? AND kind = 'free' AND status = 'paid'", quotaUser.ID).Count(&free)
		db.Model(&models.ListingPlacement{}).Where("user_id = ? AND kind = 'paid' AND status = 'pending'", quotaUser.ID).Count(&pending)
		if free != 3 || pending != 1 {
			t.Fatalf("expected 3 free and 1 pending placement, got %d and %d", free, pending)
		}
	})

	t.Run("registration requires an email code", func(t *testing.T) {
		response := requestJSON(router, http.MethodPost, "/api/v1/auth/register", "", `{"username":"email-user","email":"email-user@veloham.test","password":"secret6"}`)
		if response.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d: %s", response.Code, response.Body.String())
		}
		var started struct {
			VerificationID string `json:"verification_id"`
			DevCode        string `json:"dev_code"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &started); err != nil {
			t.Fatal(err)
		}
		if started.VerificationID == "" || started.DevCode == "" {
			t.Fatalf("unexpected registration response: %s", response.Body.String())
		}
		payload := fmt.Sprintf(`{"verification_id":%q,"code":%q}`, started.VerificationID, started.DevCode)
		response = requestJSON(router, http.MethodPost, "/api/v1/auth/register/verify", "", payload)
		if response.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", response.Code, response.Body.String())
		}
		var registered models.User
		if err := db.Where("email = ?", "email-user@veloham.test").First(&registered).Error; err != nil {
			t.Fatal(err)
		}
		if registered.Email != "email-user@veloham.test" {
			t.Fatalf("unexpected registered email: %s", registered.Email)
		}
	})

	t.Run("password can be reset with an email code", func(t *testing.T) {
		response := requestJSON(router, http.MethodPost, "/api/v1/auth/password/forgot", "", `{"email":"integration-owner@veloham.test"}`)
		if response.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d: %s", response.Code, response.Body.String())
		}
		var started struct {
			ResetID string `json:"reset_id"`
			DevCode string `json:"dev_code"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &started); err != nil {
			t.Fatal(err)
		}
		if started.ResetID == "" || started.DevCode == "" {
			t.Fatalf("unexpected password reset response: %s", response.Body.String())
		}
		payload := fmt.Sprintf(`{"reset_id":%q,"code":%q,"password":"new-secret-6"}`, started.ResetID, started.DevCode)
		response = requestJSON(router, http.MethodPost, "/api/v1/auth/password/reset", "", payload)
		if response.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d: %s", response.Code, response.Body.String())
		}
		response = requestJSON(router, http.MethodPost, "/api/v1/auth/login", "", `{"email":"integration-owner@veloham.test","password":"new-secret-6"}`)
		if response.Code != http.StatusOK {
			t.Fatalf("expected login after reset, got %d: %s", response.Code, response.Body.String())
		}
	})

	t.Run("admin boundary is enforced", func(t *testing.T) {
		response := request(router, http.MethodGet, "/api/v1/admin/users", ownerToken)
		if response.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", response.Code)
		}
		response = request(router, http.MethodGet, "/api/v1/admin/users", adminToken)
		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
		}
	})

	t.Run("blocked user receives the block reason", func(t *testing.T) {
		response := requestJSON(router, http.MethodPatch, "/api/v1/admin/users/"+other.ID+"/block", adminToken, `{"is_blocked":true,"reason":"Нарушение правил площадки"}`)
		if response.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", response.Code, response.Body.String())
		}
		response = requestJSON(router, http.MethodPost, "/api/v1/auth/login", "", `{"email":"integration-other@veloham.test","password":"wrong"}`)
		if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "Нарушение правил площадки") {
			t.Fatalf("expected block reason on login, got %d: %s", response.Code, response.Body.String())
		}
		response = request(router, http.MethodGet, "/api/v1/favorites", otherToken)
		if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "Нарушение правил площадки") {
			t.Fatalf("expected block reason for active session, got %d: %s", response.Code, response.Body.String())
		}
		response = request(router, http.MethodGet, "/api/v1/admin/block-events", adminToken)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Нарушение правил площадки") {
			t.Fatalf("expected block audit event, got %d: %s", response.Code, response.Body.String())
		}
	})
}

func newTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	base, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect to test PostgreSQL: %v", err)
	}
	schema := fmt.Sprintf("integration_%d", time.Now().UnixNano())
	if err := base.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() { _ = base.Exec("DROP SCHEMA " + schema + " CASCADE").Error })

	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: parsed.String(), PreferSimpleProtocol: true}), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect to test schema: %v", err)
	}
	if err := migrations.Up(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return db
}

func createUser(t *testing.T, db *gorm.DB, email, role string) models.User {
	t.Helper()
	user := models.User{Username: email, Email: email, PasswordHash: "unused", Role: role}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	return user
}

type httpHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func request(handler httpHandler, method, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

func requestJSON(handler httpHandler, method, path, token, payload string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

func createListingRequest(handler httpHandler, token string, number int) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fields := map[string]string{
		"title": fmt.Sprintf("Quota listing %d", number), "description": "Integration quota check",
		"city": "Бишкек", "category": "MTB", "condition": "хорошее", "price": "30000",
	}
	for key, value := range fields {
		_ = writer.WriteField(key, value)
	}
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listings", &body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}
