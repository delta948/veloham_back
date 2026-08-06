package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func uploadContext(t *testing.T, files map[string][]byte) *gin.Context {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, content := range files {
		part, err := writer.CreateFormFile("images", name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req
	return ctx
}

func TestValidateImageUploads(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
	ctx := uploadContext(t, map[string][]byte{"bike.png": png})
	files, err := validateImageUploads(ctx)
	if err != nil || len(files) != 1 {
		t.Fatalf("validateImageUploads() files=%d error=%v", len(files), err)
	}
}

func TestValidateImageUploadsRejectsExecutableContent(t *testing.T) {
	ctx := uploadContext(t, map[string][]byte{"bike.jpg": []byte("<script>alert(1)</script>")})
	if _, err := validateImageUploads(ctx); err == nil {
		t.Fatal("validateImageUploads() accepted non-image content")
	}
}
