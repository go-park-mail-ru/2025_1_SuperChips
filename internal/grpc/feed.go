package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/feed"
)

type PinService interface {
	GetPins(page int, pageSize int) ([]domain.PinData, error)
}

type GrpcFeedHandler struct {
	gen.UnimplementedFeedServer
	usecase PinService
}

func NewGrpcFeedHandler(usecase PinService) *GrpcFeedHandler {
	return &GrpcFeedHandler{
		usecase: usecase,
	}
}

func (h *GrpcFeedHandler) GetPins(ctx context.Context, in *gen.GetPinsRequest) (*gen.GetPinsResponse, error) {
	page := in.Page
	pageSize := in.PageSize

	pins, err := h.usecase.GetPins(int(page), int(pageSize))
	if err != nil {
		return nil, err
	}

	return &gen.GetPinsResponse{
		Pins: pinsToGrpc(pins),
	}, nil
}

func pinsToGrpc(pins []domain.PinData) []*gen.Pin {
	var grpcPins []*gen.Pin
	for i := range pins {
		pin := pins[i]
		grpcPins = append(grpcPins, &gen.Pin{
			FlowId: pin.FlowID,
			Header: pin.Header,
			AuthorId: pin.AuthorID,
			AuthorUsername: pin.AuthorUsername,
			Description: pin.Description,
			MediaUrl: pin.MediaURL,
			IsPrivate: pin.IsPrivate,
			CreatedAt: pin.CreatedAt,
			UpdatedAt: pin.UpdatedAt,
			IsLiked: pin.IsLiked,
			LikeCount: int64(pin.LikeCount),
			Width: int64(pin.Width),
			Height: int64(pin.Height),
		})
	}

	return grpcPins
}

