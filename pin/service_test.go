package pin

import (
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	mocks "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/pin/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPinService_GetPins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPinRepository(ctrl)

	baseURL := "http://example.com"
	imageDir := "/images"

	service := NewPinService(mockRepo, baseURL, imageDir)

	t.Run("Success", func(t *testing.T) {
		page := 1
		pageSize := 10

		mockRepo.EXPECT().
			GetPins(page, pageSize).
			Return([]domain.PinData{
				{
					FlowID:   1,
					Header:   "Pin 1",
					MediaURL: "image1.jpg",
				},
				{
					FlowID:   2,
					Header:   "Pin 2",
					MediaURL: "image2.jpg",
				},
			}, nil)

		pins, err := service.GetPins(page, pageSize)
		assert.NoError(t, err)
		assert.Len(t, pins, 2)

		expectedURL1 := "image1.jpg"
		expectedURL2 := "image2.jpg"

		assert.Equal(t, expectedURL1, pins[0].MediaURL)
		assert.Equal(t, expectedURL2, pins[1].MediaURL)
	})

	t.Run("Error from Repository", func(t *testing.T) {
		page := 1
		pageSize := 10

		mockRepo.EXPECT().
			GetPins(page, pageSize).
			Return(nil, errors.New("database error"))

		_, err := service.GetPins(page, pageSize)
		assert.Error(t, err)
	})
}
