package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Habeebamoo/intunel-backend/internal/models"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	return emailRegex.MatchString(email)
}

func ValidateReq(n *models.Notification) error {
	if n.Channel != "email" {
		return fmt.Errorf("channel error, only 'email' is allowed")
	}
	
	if n.To == "" || n.Title == "" || n.Body == "" {
		return fmt.Errorf("to, title, and body are all required")
	}

	if !IsValidEmail(n.To) {
		return fmt.Errorf("invalid recipient email address")
	}

	return nil
}