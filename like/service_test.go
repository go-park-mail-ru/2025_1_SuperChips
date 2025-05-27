package like

import (
    "context"
    "errors"
    "testing"

    "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
    mock_like "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/like/repository"
    "github.com/stretchr/testify/assert"
    "go.uber.org/mock/gomock"
)

func TestLikeFlow_InvalidIDs(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mock_like.NewMockLikeRepository(ctrl)
    mockPinRepo := mock_like.NewMockPinRepository(ctrl)
    service := NewLikeService(mockRepo, mockPinRepo)

    action, _, err := service.LikeFlow(context.Background(), 0, 1)
    assert.Error(t, err)
    assert.Empty(t, action)
    assert.Equal(t, "id cannot be less than or equal to zero", err.Error())

    action, _, err = service.LikeFlow(context.Background(), 1, -1)
    assert.Error(t, err)
    assert.Empty(t, action)
    assert.Equal(t, "id cannot be less than or equal to zero", err.Error())
}

func TestLikeFlow_SuccessfulLike(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mock_like.NewMockLikeRepository(ctrl)
    mockPinRepo := mock_like.NewMockPinRepository(ctrl)
    service := NewLikeService(mockRepo, mockPinRepo)

    mockPinRepo.EXPECT().
        GetPin(gomock.Any(), uint64(1), uint64(2)).
        Return(domain.PinData{IsPrivate: false}, uint64(1), nil)

    mockRepo.EXPECT().
        LikeFlow(gomock.Any(), 1, 2).
        Return("insert", "testuser", nil)

    action, username, err := service.LikeFlow(context.Background(), 1, 2)
    assert.NoError(t, err)
    assert.Equal(t, "liked", action)
    assert.Equal(t, "testuser", username)
}

func TestLikeFlow_SuccessfulUnlike(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mock_like.NewMockLikeRepository(ctrl)
    mockPinRepo := mock_like.NewMockPinRepository(ctrl)
    service := NewLikeService(mockRepo, mockPinRepo)

    mockPinRepo.EXPECT().
        GetPin(gomock.Any(), uint64(1), uint64(2)).
        Return(domain.PinData{IsPrivate: false}, uint64(1), nil)

    mockRepo.EXPECT().
        LikeFlow(gomock.Any(), 1, 2).
        Return("delete", "testuser", nil)

    action, username, err := service.LikeFlow(context.Background(), 1, 2)
    assert.NoError(t, err)
    assert.Equal(t, "unliked", action)
    assert.Equal(t, "testuser", username)
}

func TestLikeFlow_RepositoryError(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mock_like.NewMockLikeRepository(ctrl)
    mockPinRepo := mock_like.NewMockPinRepository(ctrl)
    service := NewLikeService(mockRepo, mockPinRepo)

    mockPinRepo.EXPECT().
        GetPin(gomock.Any(), uint64(1), uint64(2)).
        Return(domain.PinData{IsPrivate: false}, uint64(1), nil)

    mockRepo.EXPECT().
        LikeFlow(gomock.Any(), 1, 2).
        Return("", "", domain.ErrForbidden)

    action, username, err := service.LikeFlow(context.Background(), 1, 2)
    assert.Error(t, err)
    assert.Equal(t, domain.ErrForbidden, err)
    assert.Empty(t, action)
    assert.Empty(t, username)
}

func TestLikeFlow_UnexpectedError(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mock_like.NewMockLikeRepository(ctrl)
    mockPinRepo := mock_like.NewMockPinRepository(ctrl)
    service := NewLikeService(mockRepo, mockPinRepo)

    mockPinRepo.EXPECT().
        GetPin(gomock.Any(), uint64(1), uint64(2)).
        Return(domain.PinData{IsPrivate: false}, uint64(1), nil)

    mockRepo.EXPECT().
        LikeFlow(gomock.Any(), 1, 2).
        Return("", "", errors.New("database error"))

    action, username, err := service.LikeFlow(context.Background(), 1, 2)
    assert.Error(t, err)
    assert.EqualError(t, err, "database error")
    assert.Empty(t, action)
    assert.Empty(t, username)
}

func TestLikeFlow_PrivatePin(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mock_like.NewMockLikeRepository(ctrl)
    mockPinRepo := mock_like.NewMockPinRepository(ctrl)
    service := NewLikeService(mockRepo, mockPinRepo)

    mockPinRepo.EXPECT().
        GetPin(gomock.Any(), uint64(1), uint64(2)).
        Return(domain.PinData{IsPrivate: true}, uint64(3), nil)

    action, username, err := service.LikeFlow(context.Background(), 1, 2)
    assert.Error(t, err)
    assert.Equal(t, domain.ErrForbidden, err)
    assert.Empty(t, action)
    assert.Empty(t, username)
}