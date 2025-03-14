package id

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	uuid := GenerateUUID()
	assert.True(t, strings.Contains(uuid, "-"))
}

func TestGenerateUUIDWithDash(t *testing.T) {
	uuid := GenerateUUID(true)
	assert.False(t, strings.Contains(uuid, "-"))
}

func TestGenerateUUIDWithNoDash(t *testing.T) {
	uuid := GenerateUUID(false)
	assert.True(t, strings.Contains(uuid, "-"))
}
