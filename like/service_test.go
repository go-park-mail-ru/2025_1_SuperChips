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
    service := NewLikeService(mockRepo)

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
    service := NewLikeService(mockRepo)

    mockRepo.EXPECT().LikeFlow(gomock.Any(), 1, 2).Return("insert", nil)

    action, _, err := service.LikeFlow(context.Background(), 1, 2)
    assert.NoError(t, err)
    assert.Equal(t, "liked", action)
}

func TestLikeFlow_SuccessfulUnlike(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mock_like.NewMockLikeRepository(ctrl)
    service := NewLikeService(mockRepo)

    mockRepo.EXPECT().LikeFlow(gomock.Any(), 1, 2).Return("delete", nil)

    action, _, err := service.LikeFlow(context.Background(), 1, 2)
    assert.NoError(t, err)
    assert.Equal(t, "unliked", action)
}

func TestLikeFlow_RepositoryError(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mock_like.NewMockLikeRepository(ctrl)
    service := NewLikeService(mockRepo)

    mockRepo.EXPECT().LikeFlow(gomock.Any(), 1, 2).Return("", domain.ErrForbidden)

    action, _, err := service.LikeFlow(context.Background(), 1, 2)
    assert.Error(t, err)
    assert.Equal(t, domain.ErrForbidden, err)
    assert.Empty(t, action)
}

func TestLikeFlow_UnexpectedError(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mock_like.NewMockLikeRepository(ctrl)
    service := NewLikeService(mockRepo)

    mockRepo.EXPECT().LikeFlow(gomock.Any(), 1, 2).Return("", errors.New("database error"))

    action, _, err := service.LikeFlow(context.Background(), 1, 2)
    assert.Error(t, err)
    assert.EqualError(t, err, "database error")
    assert.Empty(t, action)
}