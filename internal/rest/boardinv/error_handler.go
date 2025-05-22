package rest

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/board"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
)

func handleBoardInvError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, board.ErrForbidden):
		rest.HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	default:
		rest.HttpErrorToJson(w, err.Error(), http.StatusInternalServerError)
		return
	}
}