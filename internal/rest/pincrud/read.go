package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
)

// ReadHandler godoc
//	@Summary		Get public pin by ID or private pin if user its author
//	@Description	Returns Pin Data
//	@Produce		json
//	@Param			id	query	int							true	"requested pin"
//	@Success		200	string	serverResponse.Data			"OK"
//	@Failure		400	string	serverResponse.Description	"invalid query parameter [id]"
//	@Failure		403	string	serverResponse.Description	"access to private pin is forbidden"
//	@Failure		404	string	serverResponse.Description	"no pin with given id"
//	@Failure		500	string	serverResponse.Description	"untracked error: ${error}"
//	@Router			/api/v1/flows [get]
func (app PinCRUDHandler) ReadHandler(w http.ResponseWriter, r *http.Request) {
	claims, isAuthorized := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	pinID, err := parsePinID(r.URL.Query().Get("id"))
	if err != nil {
		rest.HttpErrorToJson(w, "invalid query parameter [id]", http.StatusBadRequest)
		return
	}

	var data domain.PinData
	if isAuthorized {
		data, err = app.PinService.GetAnyPin(r.Context(), pinID, uint64(claims.UserID))
	} else {
		data, err = app.PinService.GetPublicPin(r.Context(), pinID)
	}
	if errors.Is(err, pincrud.ErrForbidden) {
		rest.HttpErrorToJson(w, "access to private pin is forbidden", http.StatusForbidden)
		return
	}
	if errors.Is(err, pincrud.ErrPinNotFound) {
		rest.HttpErrorToJson(w, "no pin with given id", http.StatusNotFound)
		return
	}
	if err != nil {
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userAgent := r.Header.Get("User-Agent")
	botUserAgents := []string{
		"facebookexternalhit",
		"Twitterbot",
		"TelegramBot",
		"Slackbot",
		"WhatsApp",
		"vkShare",
		"LinkedInBot",
		"Discordbot",
		"Googlebot",
	}

	isBot := false
	for _, botAgent := range botUserAgents {
		if strings.Contains(userAgent, botAgent) {
			isBot = true
			break
		}
	}

	var html string

	if isBot {
		html = fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8" /><title></title>
	<meta property="og:title" content="Flow" />
	<meta property="og:type" content="website" />
	<meta property="og:image" content="%s" />
	<meta property="og:url" content="https://yourflow.ru/flow/%d" />
</head></html>`, data.MediaURL, pinID)
		w.Write([]byte(html))
		return
	}

	response := rest.ServerResponse{
		Description: "OK",
		Data:        data,
	}
	rest.ServerGenerateJSONResponse(w, response, http.StatusOK)
}
