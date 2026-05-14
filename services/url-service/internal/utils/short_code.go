package utils

import "crypto/rand"

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortCode(length int) (string, error) {
	b := make([]byte, length)
	
	_, err := rand.Read(b)
	if err != nil {
		return  "", err
	}
	
	for i := range b {
		b[i] = charset[int(b[i]) % len(charset)]
	}
	
	return string(b), nil
}

func IsValidShortCode(code string) bool {
	if len(code) < 3 || len(code) > 20 {
		return false
	}
	for _, char := range code {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}
	return true
}

func NewError(msg string) error {
	return &CustomError{Message: msg}
}

type CustomError struct {
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}