package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	mocks "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/chat/grpc"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/chat"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetChats(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockChatService := mocks.NewMockChatServiceClient(ctrl)

    handler := ChatHandler{
        ChatService:       mockChatService,
        ContextExpiration: time.Second,
    }

    t.Run("Success", func(t *testing.T) {
        mockChatService.EXPECT().
            GetChats(gomock.Any(), &gen.GetChatsRequest{Username: "test_user"}).
            Return(&gen.ChatsStruct{
                Chats: []*gen.Chat{
                    {
                        ChatID:       1,
                        Username:     "user1",
                        PublicName:   "User One",
                        Avatar:       "avatar1.jpg",
                        MessageCount: 5,
                        Messages: &gen.MessagesStruct{
                            Messages: []*gen.Message{
                                {
                                    MessageID: 101,
                                    Content:   "Hello!",
                                    Sender:    "user1",
                                    Timestamp: timestamppb.Now(),
                                    IsRead:    true,
                                },
                            },
                        },
                    },
                },
            }, nil)

        req := httptest.NewRequest(http.MethodGet, "/api/v1/chats", nil)
        ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{Username: "test_user"})
        req = req.WithContext(ctx)

        rr := httptest.NewRecorder()

        handler.GetChats(rr, req)

        assert.Equal(t, http.StatusOK, rr.Code)
        assert.Contains(t, rr.Body.String(), `"username":"user1"`)
        assert.Contains(t, rr.Body.String(), `"public_name":"User One"`)
    })

    t.Run("Error from ChatService", func(t *testing.T) {
        mockChatService.EXPECT().
            GetChats(gomock.Any(), &gen.GetChatsRequest{Username: "test_user"}).
            Return(nil, errors.New("database error"))

        req := httptest.NewRequest(http.MethodGet, "/api/v1/chats", nil)
        ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{Username: "test_user"})
        req = req.WithContext(ctx)

        rr := httptest.NewRecorder()

        handler.GetChats(rr, req)

        assert.Equal(t, http.StatusInternalServerError, rr.Code)
        assert.Contains(t, rr.Body.String(), "Internal Server Error")
    })
}

func TestNewChat(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockChatService := mocks.NewMockChatServiceClient(ctrl)

    handler := ChatHandler{
        ChatService:       mockChatService,
        ContextExpiration: time.Second,
    }

    t.Run("Success", func(t *testing.T) {
        targetUser := Username{Username: "target_user"}

        body, _ := json.Marshal(targetUser)

        mockChatService.EXPECT().
            CreateChat(gomock.Any(), &gen.CreateChatRequest{
                Username:       "test_user",
                TargetUsername: "target_user",
            }).
            Return(&gen.CreateChatResponse{
                Chat: &gen.Chat{
                    ChatID:       1,
                    Username:     "target_user",
                    PublicName:   "Target User",
                    Avatar:       "avatar.jpg",
                    MessageCount: 0,
					Messages: &gen.MessagesStruct{
						Messages: nil,
					},
                },
            }, nil)

        req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/new", bytes.NewBuffer(body))
        ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{Username: "test_user"})
        req = req.WithContext(ctx)

        rr := httptest.NewRecorder()

        handler.NewChat(rr, req)

        assert.Equal(t, http.StatusCreated, rr.Code)
        assert.Contains(t, rr.Body.String(), `"username":"target_user"`)
        assert.Contains(t, rr.Body.String(), `"public_name":"Target User"`)
    })

    t.Run("Validation Error - Missing Target User", func(t *testing.T) {
        body, _ := json.Marshal(Username{})

        req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/new", bytes.NewBuffer(body))
        ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{Username: "test_user"})
        req = req.WithContext(ctx)

        rr := httptest.NewRecorder()

        handler.NewChat(rr, req)

        assert.Equal(t, http.StatusBadRequest, rr.Code)
        assert.Contains(t, rr.Body.String(), "Bad Request")
    })
}

func TestGetContacts(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockChatService := mocks.NewMockChatServiceClient(ctrl)

    handler := ChatHandler{
        ChatService:       mockChatService,
        ContextExpiration: time.Second,
    }

    t.Run("Success", func(t *testing.T) {
        mockChatService.EXPECT().
            GetContacts(gomock.Any(), &gen.GetContactsRequest{Username: "test_user"}).
            Return(&gen.ContactsStruct{
                Contacts: []*gen.Contact{
                    {
                        Username:       "contact1",
                        PublicUsername: "Contact One",
                        Avatar:         "avatar1.jpg",
                    },
                    {
                        Username:       "contact2",
                        PublicUsername: "Contact Two",
                        Avatar:         "avatar2.jpg",
                    },
                },
            }, nil)

        req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts", nil)
        ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{Username: "test_user"})
        req = req.WithContext(ctx)

        rr := httptest.NewRecorder()

        handler.GetContacts(rr, req)

        assert.Equal(t, http.StatusOK, rr.Code)
        assert.Contains(t, rr.Body.String(), `"username":"contact1"`)
        assert.Contains(t, rr.Body.String(), `"public_name":"Contact One"`)
    })

    t.Run("Error from ChatService", func(t *testing.T) {
        mockChatService.EXPECT().
            GetContacts(gomock.Any(), &gen.GetContactsRequest{Username: "test_user"}).
            Return(nil, errors.New("database error"))

        req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts", nil)
        ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{Username: "test_user"})
        req = req.WithContext(ctx)

        rr := httptest.NewRecorder()

        handler.GetContacts(rr, req)

        assert.Equal(t, http.StatusInternalServerError, rr.Code)
        assert.Contains(t, rr.Body.String(), "Internal Server Error")
    })
}

func TestCreateContact(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockChatService := mocks.NewMockChatServiceClient(ctrl)

    handler := ChatHandler{
        ChatService:       mockChatService,
        ContextExpiration: time.Second,
    }

    t.Run("Success", func(t *testing.T) {
        user := Username{Username: "new_contact"}

        body, _ := json.Marshal(user)

        mockChatService.EXPECT().
            CreateContact(gomock.Any(), &gen.CreateContactRequest{
                Username:       "test_user",
                TargetUsername: "new_contact",
            }).
            Return(&gen.CreateContactResponse{
                ChatID:     1,
                Avatar:     "avatar.jpg",
                PublicName: "New Contact",
            }, nil)

        req := httptest.NewRequest(http.MethodPost, "/api/v1/contact", bytes.NewBuffer(body))
        ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{Username: "test_user"})
        req = req.WithContext(ctx)

        rr := httptest.NewRecorder()

        handler.CreateContact(rr, req)

        assert.Equal(t, http.StatusCreated, rr.Code)
        assert.Contains(t, rr.Body.String(), `"chat_id":1`)
        assert.Contains(t, rr.Body.String(), `"avatar":"avatar.jpg"`)
    })

    t.Run("Validation Error - Missing Username", func(t *testing.T) {
        body, _ := json.Marshal(Username{})

        req := httptest.NewRequest(http.MethodPost, "/api/v1/contact", bytes.NewBuffer(body))
        ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{Username: "test_user"})
        req = req.WithContext(ctx)

        rr := httptest.NewRecorder()

        handler.CreateContact(rr, req)

        assert.Equal(t, http.StatusBadRequest, rr.Code)
        assert.Contains(t, rr.Body.String(), "Bad Request")
    })
}

func TestGetChat(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockChatService := mocks.NewMockChatServiceClient(ctrl)

    handler := ChatHandler{
        ChatService:       mockChatService,
        ContextExpiration: time.Second,
    }

    t.Run("Success", func(t *testing.T) {
        mockChatService.EXPECT().
            GetChat(gomock.Any(), &gen.GetChatRequest{
                ChatID:   1,
                Username: "test_user",
            }).
            Return(&gen.Chat{
                ChatID:       1,
                Username:     "user1",
                PublicName:   "User One",
                Avatar:       "avatar.jpg",
                MessageCount: 3,
                Messages: &gen.MessagesStruct{
                    Messages: []*gen.Message{
                        {
                            MessageID: 101,
                            Content:   "Hello!",
                            Sender:    "user1",
                            Timestamp: timestamppb.Now(),
                            IsRead:    true,
                        },
                    },
                },
            }, nil)

        req := httptest.NewRequest(http.MethodGet, "/api/v1/chat?id=1", nil)
        ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{Username: "test_user"})
        req = req.WithContext(ctx)

        rr := httptest.NewRecorder()

        handler.GetChat(rr, req)

        assert.Equal(t, http.StatusOK, rr.Code)
        assert.Contains(t, rr.Body.String(), `"username":"user1"`)
        assert.Contains(t, rr.Body.String(), `"public_name":"User One"`)
    })

    t.Run("Invalid ID", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/api/v1/chat?id=invalid", nil)
        ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{Username: "test_user"})
        req = req.WithContext(ctx)

        rr := httptest.NewRecorder()

        handler.GetChat(rr, req)

        assert.Equal(t, http.StatusBadRequest, rr.Code)
        assert.Contains(t, rr.Body.String(), "id must be an integer")
    })
}
