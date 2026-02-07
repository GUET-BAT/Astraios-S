package logic

import (
	"fmt"
	"unicode"
)

// Input validation constraints.
const (
	MinUsernameLength = 3
	MaxUsernameLength = 32
	MinPasswordLength = 8
	MaxPasswordLength = 72 // bcrypt max input length
)

func validateUsername(username string) error {
	if len(username) < MinUsernameLength || len(username) > MaxUsernameLength {
		return fmt.Errorf("username length must be %d-%d characters", MinUsernameLength, MaxUsernameLength)
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	if len(password) > MaxPasswordLength {
		return fmt.Errorf("password must be at most %d characters", MaxPasswordLength)
	}
	var hasLetter, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsLetter(c):
			hasLetter = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return fmt.Errorf("password must contain both letters and digits")
	}
	return nil
}
