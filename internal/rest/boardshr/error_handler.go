package rest

import (
	"errors"
	"net/http"

	boardshrService "github.com/go-park-mail-ru/2025_1_SuperChips/boardshr"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
)

func handleBoardShrError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	case errors.Is(err, boardshrService.ErrLinkNotFound):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	case errors.Is(err, boardshrService.ErrNonExistentUsername):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	case errors.Is(err, boardshrService.ErrInconsistentDataInDB):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	case errors.Is(err, boardshrService.ErrAlreadyEditor):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	case errors.Is(err, boardshrService.ErrForbbiden):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	case errors.Is(err, boardshrService.ErrLinkExpired):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusGone), http.StatusGone)
		return
	case errors.Is(err, boardshrService.ErrFailCoauthorInsert):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	case errors.Is(err, boardshrService.ErrFailCoauthorDelete):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	case errors.Is(err, boardshrService.ErrAuthorRefuseEditing):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	default:
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}