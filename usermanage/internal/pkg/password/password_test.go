package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		expectError bool
	}{
		{"Valid length", 12, false},
		{"Valid longer length", 20, false},
		{"Invalid shorter length", 8, true},
		{"Invalid zero length", 0, true},
		{"Invalid negative length", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GeneratePassword(tt.length)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, password)
			} else {
				assert.NoError(t, err)
				assert.Len(t, password, tt.length)

				// Check if password contains required character types
				assert.Regexp(t, "[a-z]", password, "password should contain lowercase letters")
				assert.Regexp(t, "[A-Z]", password, "password should contain uppercase letters")
				assert.Regexp(t, "[0-9]", password, "password should contain digits")
				assert.Regexp(t, "[!@#$%^&*\\-=]", password, "password should contain symbols")
			}
		})
	}
}

func TestHashAndVerify(t *testing.T) {
	password := "SecureP@ssword123"

	hash, err := Hash(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Valid password verification
	assert.True(t, Verify(hash, password))

	// Invalid password verification
	assert.False(t, Verify(hash, "WrongPassword123"))
}

func TestDefaultStrengthOptions(t *testing.T) {
	opts := DefaultStrengthOptions()

	assert.Equal(t, 8, opts.MinLength)
	assert.Equal(t, 32, opts.MaxLength)
	assert.Equal(t, 1, opts.MinUpperCase)
	assert.Equal(t, 1, opts.MinLowerCase)
	assert.Equal(t, 1, opts.MinDigits)
	assert.Equal(t, 1, opts.MinSpecial)
	assert.True(t, opts.RequireUnique)
}
