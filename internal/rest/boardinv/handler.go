package rest

import (
	"time"
)

type BoardInvHandler struct {
	BoardInvService BoardInvServicer
	ContextDeadline time.Duration
}
