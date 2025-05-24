package rest

import (
	"time"
)

type BoardShrHandler struct {
	BoardShrService BoardShrServicer
	ContextDeadline time.Duration
}
