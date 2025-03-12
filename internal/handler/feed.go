package handler

import (
	"net/http"
	"strconv"
)

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
func (app AppHandler) FeedHandler(w http.ResponseWriter, r *http.Request) {
    pageSize := app.Config.PageSize
    page := parsePageQueryParam(r.URL.Query().Get("page"))

    if page < 1 {
		handleHttpError(w, "bad request", http.StatusBadRequest)
        return
    }

    pagedImages := app.PinStorage.GetPinPage(page, pageSize)
    if len(pagedImages) == 0 {
        handleHttpError(w, "page not found", http.StatusNotFound)
        return
    }

    response := serverResponse{
		Data: pagedImages,
	}

	serverGenerateJSONResponse(w, response, http.StatusOK)
}

func parsePageQueryParam(pageStr string) int {
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        return 1
    }
    return page
}

