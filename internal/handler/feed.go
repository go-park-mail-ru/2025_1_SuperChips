package handler

import (
	"net/http"
	"strconv"
)

func parsePageQueryParam(pageStr string) int {
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        return 1
    }
    return page
}

func (app AppHandler) FeedHandler(w http.ResponseWriter, r *http.Request) {
	imageFiles := app.PinStorage.Pins

    pageSize := 10
    page := parsePageQueryParam(r.URL.Query().Get("page"))
	if page == 0 {

	}

    totalPages := (len(imageFiles) + pageSize - 1) / pageSize

    if page < 1 || page > totalPages {
		handleHttpError(w, "Invalid page number", http.StatusBadRequest)
        return
    }

    startIndex := (page - 1) * pageSize
    endIndex := startIndex + pageSize
    if endIndex > len(imageFiles) {
        endIndex = len(imageFiles)
    }

    pagedImages := imageFiles[startIndex:endIndex]

    response := serverResponse{
		Data: pagedImages,
	}

	serverGenerateJSONResponse(w, response, http.StatusOK)
}

