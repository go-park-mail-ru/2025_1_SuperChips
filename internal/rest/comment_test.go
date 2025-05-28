package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCommentService struct {
	mock.Mock
}

func (m *MockCommentService) GetComments(ctx context.Context, flowID, userID, page, size int) ([]domain.Comment, error) {
	args := m.Called(ctx, flowID, userID, page, size)
	return args.Get(0).([]domain.Comment), args.Error(1)
}

func (m *MockCommentService) LikeComment(ctx context.Context, flowID, commentID, userID int) (string, error) {
	args := m.Called(ctx, flowID, commentID, userID)
	return args.String(0), args.Error(1)
}

func (m *MockCommentService) AddComment(ctx context.Context, flowID, userID int, content string) error {
	args := m.Called(ctx, flowID, userID, content)
	return args.Error(0)
}

func (m *MockCommentService) DeleteComment(ctx context.Context, commentID, userID int) error {
	args := m.Called(ctx, commentID, userID)
	return args.Error(0)
}

func TestCommentHandler_GetComments(t *testing.T) {
    tests := []struct {
        name           string
        url            string
        userID         int
        mockComments   []domain.Comment
        mockError      error
        expectedStatus int
    }{
        {
            name:   "Success",
            url:    "/flows/1/comments?page=1&size=10",
            userID: 2,
            mockComments: []domain.Comment{
                {ID: 1, AuthorID: 2, FlowID: 1, Content: "Test comment"},
            },
            expectedStatus: http.StatusOK,
        },
        {
            name:           "Invalid flow ID",
            url:            "/flows/invalid/comments",
            expectedStatus: http.StatusBadRequest,
        },
        {
            name:           "Invalid page",
            url:            "/flows/1/comments?page=invalid&size=10",
            expectedStatus: http.StatusBadRequest,
        },
        {
            name:           "Invalid size",
            url:            "/flows/1/comments?page=1&size=invalid",
            expectedStatus: http.StatusBadRequest,
        },
        {
            name:           "Not found",
            url:            "/flows/1/comments?page=1&size=10",
            userID:         2,
            mockComments:   []domain.Comment{},
            mockError:      nil,
            expectedStatus: http.StatusNotFound,
        },
        {
            name:           "Forbidden",
            url:            "/flows/1/comments?page=1&size=10",
            userID:         2,
            mockComments:   nil,
            mockError:      domain.ErrForbidden,
            expectedStatus: http.StatusForbidden,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockService := new(MockCommentService)
            handler := CommentHandler{
                Service:           mockService,
                ContextExpiration: time.Second,
            }

            // Set up mock expectations before making the request
            if tt.mockComments != nil || tt.mockError != nil {
                flowID := 1 // Default for valid cases
                if tt.name == "Invalid flow ID" {
                    // Skip mock setup for invalid flow ID case
                } else {
                    page := 1
                    size := 10
                    if tt.name == "Invalid page" || tt.name == "Invalid size" {
                        // Skip mock setup for invalid pagination cases
                    } else {
                        mockService.On("GetComments", mock.Anything, flowID, tt.userID, page, size).
                            Return(tt.mockComments, tt.mockError)
                    }
                }
            }

            req := httptest.NewRequest(http.MethodGet, tt.url, nil)
            if tt.userID != 0 {
                req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: tt.userID}))
            }

            w := httptest.NewRecorder()

            // Create a router to handle path parameters
            router := http.NewServeMux()
            router.HandleFunc("GET /flows/{flow_id}/comments", handler.GetComments)
            router.ServeHTTP(w, req)

            assert.Equal(t, tt.expectedStatus, w.Code)

            if tt.expectedStatus == http.StatusOK {
                var response ServerResponse
                err := json.Unmarshal(w.Body.Bytes(), &response)
                assert.NoError(t, err)
                assert.Equal(t, "OK", response.Description)
                assert.NotNil(t, response.Data)
            }

            if tt.mockComments != nil || tt.mockError != nil {
                mockService.AssertExpectations(t)
            }
        })
    }
}

func TestCommentHandler_LikeComment(t *testing.T) {
    tests := []struct {
        name           string
        url            string
        userID         int
        mockAction     string
        mockError      error
        expectedStatus int
    }{
        {
            name:           "Success like",
            url:            "/flows/1/comments/1/like",
            userID:         2,
            mockAction:     "insert",
            expectedStatus: http.StatusOK,
        },
        {
            name:           "Success unlike",
            url:            "/flows/1/comments/1/like",
            userID:         2,
            mockAction:     "delete",
            expectedStatus: http.StatusOK,
        },
        {
            name:           "Invalid flow ID",
            url:            "/flows/invalid/comments/1/like",
            expectedStatus: http.StatusBadRequest,
        },
        {
            name:           "Invalid comment ID",
            url:            "/flows/1/comments/invalid/like",
            expectedStatus: http.StatusBadRequest,
        },
        {
            name:           "Forbidden",
            url:            "/flows/1/comments/1/like",
            userID:         2,
            mockAction:     "",
            mockError:      domain.ErrForbidden,
            expectedStatus: http.StatusForbidden,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockService := new(MockCommentService)
            handler := CommentHandler{
                Service:           mockService,
                ContextExpiration: time.Second,
            }

            // Only set up mock expectations for valid IDs
            parts := strings.Split(tt.url, "/")
            if len(parts) >= 4 {
                flowID, err1 := strconv.Atoi(parts[2])
                commentID, err2 := strconv.Atoi(parts[4])
                if err1 == nil && err2 == nil {
                    mockService.On("LikeComment", mock.Anything, flowID, commentID, tt.userID).
                        Return(tt.mockAction, tt.mockError)
                }
            }

            req := httptest.NewRequest(http.MethodPost, tt.url, nil)
            if tt.userID != 0 {
                req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: tt.userID}))
            }

            w := httptest.NewRecorder()

            // Create a router to handle path parameters
            router := http.NewServeMux()
            router.HandleFunc("POST /flows/{flow_id}/comments/{comment_id}/like", handler.LikeComment)
            router.ServeHTTP(w, req)

            assert.Equal(t, tt.expectedStatus, w.Code)

            if tt.expectedStatus == http.StatusOK {
                var response ServerResponse
                err := json.Unmarshal(w.Body.Bytes(), &response)
                assert.NoError(t, err)
                assert.Equal(t, "OK", response.Description)

                var action struct {
                    Action string `json:"action"`
                }
                dataBytes, _ := json.Marshal(response.Data)
                json.Unmarshal(dataBytes, &action)
                assert.Equal(t, tt.mockAction, action.Action)
            }

            if tt.url != "/flows/invalid/comments/1/like" && tt.url != "/flows/1/comments/invalid/like" {
                mockService.AssertExpectations(t)
            }
        })
    }
}

func TestCommentHandler_AddComment(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		userID         int
		comment        domain.Comment
		mockError      error
		expectedStatus int
	}{
		{
			name:   "Success",
			url:    "/flows/1/comments",
			userID: 2,
			comment: domain.Comment{
				Content: "Valid comment",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Invalid flow ID",
			url:    "/flows/invalid/comments",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Empty content",
			url:    "/flows/1/comments",
			userID: 2,
			comment: domain.Comment{
				Content: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Forbidden",
			url:    "/flows/1/comments",
			userID: 2,
			comment: domain.Comment{
				Content: "Test comment",
			},
			mockError:      domain.ErrForbidden,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockCommentService)
			handler := CommentHandler{
				Service:           mockService,
				ContextExpiration: time.Second,
			}

			body, _ := json.Marshal(tt.comment)
			req := httptest.NewRequest(http.MethodPost, tt.url, bytes.NewReader(body))
			if tt.userID != 0 {
				req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: tt.userID}))
			}

			// Extract flowID from URL for mock setup if it's a valid number
			flowIDStr := req.URL.Path[len("/flows/"):][:len(req.URL.Path[len("/flows/"):])-len("/comments")]
			if flowID, err := strconv.Atoi(flowIDStr); err == nil && tt.comment.Content != "" {
				mockService.On("AddComment", mock.Anything, flowID, tt.userID, tt.comment.Content).
					Return(tt.mockError)
			}

			w := httptest.NewRecorder()

			// Create a router to handle path parameters
			router := http.NewServeMux()
			router.HandleFunc("POST /flows/{flow_id}/comments", handler.AddComment)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response ServerResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Created", response.Description)
			}

			if tt.comment.Content != "" || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestCommentHandler_DeleteComment(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		userID         int
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Success",
			url:            "/comments/1",
			userID:         2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid comment ID",
			url:            "/comments/invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Forbidden",
			url:            "/comments/1",
			userID:         2,
			mockError:      domain.ErrForbidden,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockCommentService)
			handler := CommentHandler{
				Service:           mockService,
				ContextExpiration: time.Second,
			}

			req := httptest.NewRequest(http.MethodDelete, tt.url, nil)
			if tt.userID != 0 {
				req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: tt.userID}))
			}

			// Extract commentID from URL for mock setup if it's a valid number
			commentIDStr := req.URL.Path[len("/comments/"):]
			if commentID, err := strconv.Atoi(commentIDStr); err == nil {
				mockService.On("DeleteComment", mock.Anything, commentID, tt.userID).
					Return(tt.mockError)
			}

			w := httptest.NewRecorder()

			// Create a router to handle path parameters
			router := http.NewServeMux()
			router.HandleFunc("DELETE /comments/{comment_id}", handler.DeleteComment)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response ServerResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Deleted", response.Description)
			}

			if tt.url != "/comments/invalid" {
				mockService.AssertExpectations(t)
			}
		})
	}
}