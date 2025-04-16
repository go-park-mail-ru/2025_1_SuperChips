package rest

import "github.com/go-park-mail-ru/2025_1_SuperChips/configs"

type PinCRUDHandler struct {
	Config     configs.Config
	PinService PinCRUDServicer
}
