package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	mocks "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/feed/service"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/feed"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGrpcFeedHandler_GetPins(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockPinService := mocks.NewMockPinService(ctrl)

    handler := NewGrpcFeedHandler(mockPinService)

    t.Run("Success", func(t *testing.T) {
        page := int64(1)
        pageSize := int64(10)

        mockPinService.EXPECT().
            GetPins(int(page), int(pageSize)).
            Return([]domain.PinData{
                {
                    FlowID:         1,
                    Header:         "Pin 1",
                    AuthorID:       101,
                    AuthorUsername: "user1",
                    Description:    "Description 1",
                    MediaURL:       "image1.jpg",
                    IsPrivate:      false,
                    CreatedAt:      "2023-01-01T10:00:00Z",
                    UpdatedAt:      "2023-01-01T10:00:00Z",
                    IsLiked:        true,
                    LikeCount:      5,
                    Width:          800,
                    Height:         600,
                },
                {
                    FlowID:         2,
                    Header:         "Pin 2",
                    AuthorID:       102,
                    AuthorUsername: "user2",
                    Description:    "Description 2",
                    MediaURL:       "image2.jpg",
                    IsPrivate:      false,
                    CreatedAt:      "2023-01-02T10:00:00Z",
                    UpdatedAt:      "2023-01-02T10:00:00Z",
                    IsLiked:        false,
                    LikeCount:      3,
                    Width:          1024,
                    Height:         768,
                },
            }, nil)

        req := &gen.GetPinsRequest{
            Page:     page,
            PageSize: pageSize,
        }

        resp, err := handler.GetPins(context.Background(), req)
        assert.NoError(t, err)
        assert.NotNil(t, resp)

        assert.Len(t, resp.Pins, 2)
        assert.Equal(t, uint64(1), resp.Pins[0].FlowId)
        assert.Equal(t, "Pin 1", resp.Pins[0].Header)
        assert.Equal(t, uint64(101), resp.Pins[0].AuthorId)
        assert.Equal(t, "user1", resp.Pins[0].AuthorUsername)
        assert.Equal(t, "image1.jpg", resp.Pins[0].MediaUrl)
        assert.Equal(t, int64(5), resp.Pins[0].LikeCount)
    })

    t.Run("Error from PinService", func(t *testing.T) {
        page := int64(1)
        pageSize := int64(10)

        mockPinService.EXPECT().
            GetPins(int(page), int(pageSize)).
            Return(nil, errors.New("database error"))

        req := &gen.GetPinsRequest{
            Page:     page,
            PageSize: pageSize,
        }

        _, err := handler.GetPins(context.Background(), req)
        assert.Error(t, err)
    })
}