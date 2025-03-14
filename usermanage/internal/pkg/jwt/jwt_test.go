package jwt

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTokenEmptyUsername(t *testing.T) {
	Initialize([]byte("foo"), 2*time.Hour)
	token, _, err := GenerateToken("")
	if errors.Is(err, errors.New("username is required")) {
		t.Errorf("GenerateToken(\"\") = %s; want username is required", token)
	}
}

func TestGenerateToken(t *testing.T) {
	Initialize([]byte("foo"), 2*time.Hour)
	username := "foo"
	token, _, err := GenerateToken(username)
	assert.Nil(t, err)
	assert.Greater(t, len(token), 0)
}
