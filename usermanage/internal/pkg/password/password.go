package password

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	lowercase     = "abcdefghijklmnopqrstuvwxyz"
	uppercase     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers       = "0123456789"
	symbols       = "!@#$%^&*-="
	allCharacters = lowercase + uppercase + numbers + symbols
)

// Strength represents password strength level
type Strength int

const (
	Weak Strength = iota
	Medium
	Strong
)

// StrengthOptions configures password strength requirements
type StrengthOptions struct {
	MinLength     int
	MaxLength     int
	MinUpperCase  int
	MinLowerCase  int
	MinDigits     int
	MinSpecial    int
	RequireUnique bool
}

// DefaultStrengthOptions returns default password strength requirements
func DefaultStrengthOptions() StrengthOptions {
	return StrengthOptions{
		MinLength:     8,
		MaxLength:     32,
		MinUpperCase:  1,
		MinLowerCase:  1,
		MinDigits:     1,
		MinSpecial:    1,
		RequireUnique: true,
	}
}

// GeneratePassword generates a random password with specified length.
// The password includes at least one lowercase letter, one uppercase letter, one number, and one symbol.
// If the length is less than 12, the function returns an error message.
func GeneratePassword(length int) (string, error) {
	if length < 12 {
		return "", errors.New("password length must be at least 12")
	}

	// Pre-allocate the builder capacity to avoid resizing
	var password strings.Builder
	password.Grow(length)

	// Ensure all required character types are included
	password.WriteByte(lowercase[rand.Intn(len(lowercase))])
	password.WriteByte(uppercase[rand.Intn(len(uppercase))])
	password.WriteByte(numbers[rand.Intn(len(numbers))])
	password.WriteByte(symbols[rand.Intn(len(symbols))])

	// Fill the rest with random characters
	for i := 4; i < length; i++ {
		password.WriteByte(allCharacters[rand.Intn(len(allCharacters))])
	}

	// Convert to runes for proper character handling and shuffle
	runes := []rune(password.String())
	rand.Shuffle(len(runes), func(i, j int) {
		runes[i], runes[j] = runes[j], runes[i]
	})

	return string(runes), nil
}

// CheckStrength evaluates password strength based on given options
// Returns strength level and error message if any
func CheckStrength(password string, opts StrengthOptions) (Strength, string) {
	if len(password) < opts.MinLength {
		return Weak, "password too short"
	}
	if len(password) > opts.MaxLength {
		return Weak, "password too long"
	}

	var upper, lower, digit, special int
	seen := make(map[rune]bool)

	for _, char := range password {
		if opts.RequireUnique && seen[char] {
			return Weak, "password contains duplicate characters"
		}
		seen[char] = true

		switch {
		case unicode.IsUpper(char):
			upper++
		case unicode.IsLower(char):
			lower++
		case unicode.IsDigit(char):
			digit++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			special++
		}
	}

	if upper < opts.MinUpperCase {
		return Weak, "insufficient uppercase letters"
	}
	if lower < opts.MinLowerCase {
		return Weak, "insufficient lowercase letters"
	}
	if digit < opts.MinDigits {
		return Weak, "insufficient digits"
	}
	if special < opts.MinSpecial {
		return Weak, "insufficient special characters"
	}

	// Calculate strength based on complexity
	score := upper + lower + digit + special
	if score >= 10 {
		return Strong, ""
	}
	return Medium, ""
}

// Hash generates a bcrypt hash of the password using the default cost
func Hash(rawPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// Verify compares a hashed password with a raw password
func Verify(hashedPassword, rawPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(rawPassword))
	return err == nil
}
