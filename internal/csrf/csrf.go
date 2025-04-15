package csrf

import "github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"

const CSRFToken = "csrf_token"

func GenerateCSRF() (string, error) {
	return security.GenerateRandomHash()
}

