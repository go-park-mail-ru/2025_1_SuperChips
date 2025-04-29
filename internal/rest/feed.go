package rest

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/feed"
)

type PinServiceInterface interface {
	GetPins(page int, pageSize int) ([]domain.PinData, error)
}

type PinsHandler struct {
	Config            configs.Config
	FeedClient        gen.FeedClient
	ContextExpiration time.Duration
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

	ctx, cancel := context.WithTimeout(context.Background(), app.ContextExpiration)
	defer cancel()

	grpcResp, err := app.FeedClient.GetPins(ctx, &gen.GetPinsRequest{
		Page: int64(page),
		PageSize: int64(pageSize),
	})
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	pagedImages := grpcToNormal(grpcResp.Pins)

	if len(pagedImages) == 0 {
		HttpErrorToJson(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	domain.EscapeFlows(pagedImages)

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

func grpcToNormal(grpcPins []*gen.Pin) []domain.PinData {
	var pins []domain.PinData
	for i := range grpcPins {
		grpcPin := grpcPins[i]
		pins = append(pins, domain.PinData{
			FlowID: grpcPin.FlowId,
			Header: grpcPin.Header,
			AuthorID: grpcPin.AuthorId,
			AuthorUsername: grpcPin.AuthorUsername,
			Description: grpcPin.Description,
			MediaURL: grpcPin.MediaUrl,
			IsPrivate: grpcPin.IsPrivate,
			CreatedAt: grpcPin.CreatedAt,
			UpdatedAt: grpcPin.UpdatedAt,
			IsLiked: grpcPin.IsLiked,
			LikeCount: int(grpcPin.LikeCount),
			Width: int(grpcPin.Width),
			Height: int(grpcPin.Height),
		})
	}

	return pins
}
