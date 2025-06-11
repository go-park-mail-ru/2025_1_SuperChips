package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	starService "github.com/go-park-mail-ru/2025_1_SuperChips/star"
)

type StarHandler struct {
	Config          configs.Config
	ContextDeadline time.Duration
	StarService     StarServicer
}

func (h *StarHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	case errors.Is(err, starService.ErrAlreadyStar):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	case errors.Is(err, starService.ErrNoFreeStarSlots):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	case errors.Is(err, starService.ErrPinNotFound):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	case errors.Is(err, starService.ErrPinIsNotStar):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	case errors.Is(err, starService.ErrInconsistentDataInDB):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	default:
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
