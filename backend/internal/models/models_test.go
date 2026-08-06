package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUserEmailIsPrivateByDefault(t *testing.T) {
	user := User{ID: "user-1", Username: "rider", Email: "rider@example.com"}
	publicJSON, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(publicJSON), user.Email) {
		t.Fatalf("public user JSON leaked email: %s", publicJSON)
	}

	privateJSON, err := json.Marshal(UserWithEmail(user))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(privateJSON), user.Email) {
		t.Fatalf("private user JSON omitted email: %s", privateJSON)
	}
}
