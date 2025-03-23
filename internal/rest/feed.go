package rest

import (
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pin"
)

type PinsHandler struct {
    Config      configs.Config
	PinService  pin.PinService
}

// FeedHandler godoc
// @Summary Get Pins
// @Description Returns a pageSized number of pins
// @Accept json
// @Produce json
// @Param page path int true "requested page" example("?page=3")
// @Success 200 string serverResponse.Data "OK"
// @Failure 404 string serverResponse.Description "page not found"
// @Failure 400 string serverResponse.Description "bad request"
// @Router /api/v1/feed [get]
func (app PinsHandler) FeedHandler(w http.ResponseWriter, r *http.Request) {
    pageSize := app.Config.PageSize
    page := parsePageQueryParam(r.URL.Query().Get("page"))

    if page < 1 {
        HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
        return
    }

    pagedImages, err := app.PinService.GetPins(page, pageSize)
    if err != nil {
        println(err.Error())
        HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        return
    }

    if len(pagedImages) == 0 {
        HttpErrorToJson(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
    }

    response := ServerResponse{
		Data: pagedImages,
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}

func parsePageQueryParam(pageStr string) int {
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        return 1
    }
    return page
}

