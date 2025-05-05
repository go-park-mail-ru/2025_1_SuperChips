package rest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	chatWebsocket "github.com/go-park-mail-ru/2025_1_SuperChips/internal/websocket"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/chat"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	ChatService       gen.ChatServiceClient
	ContextExpiration time.Duration
}

type ChatWebsocketHandler struct {
	Hub *chatWebsocket.Hub
	ContextExpiration time.Duration
}

type Username struct {
	Username string `json:"username"`
}

// GET api/v1/chats
func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("id") != "" {
		h.GetChat(w, r)
		return
	}

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	grpcResp, err := h.ChatService.GetChats(ctx, &gen.GetChatsRequest{
		Username: claims.Username,
	})
	if err != nil {
		handleChatGrpcError(w, err)
		return
	}

	chats := chatsToNormal(grpcResp.Chats)

	resp := ServerResponse{
		Description: "OK",
		Data:        chats,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func handleChatGrpcError(w http.ResponseWriter, err error) {
	switch {
	default:
		HttpErrorToJson(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ChatHandler) NewChat(w http.ResponseWriter, r *http.Request) {

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	var target Username
	if err := DecodeData(w, r.Body, &target); err != nil {
		return
	}

	if target.Username == "" || claims.Username == target.Username {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	grpcResp, err := h.ChatService.CreateChat(ctx, &gen.CreateChatRequest{
		Username: claims.Username,
		TargetUsername: target.Username,
	})
	if err != nil {
		handleChatGrpcError(w, err)
		return
	}

	grpcChats := []*gen.Chat{
		grpcResp.Chat,
	}

	chat := chatsToNormal(grpcChats)

	resp := ServerResponse{
		Description: "Created",
		Data:        chat[0],
	}

	ServerGenerateJSONResponse(w, resp, http.StatusCreated)
}

func (h *ChatHandler) GetContacts(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	grpcResp, err := h.ChatService.GetContacts(ctx, &gen.GetContactsRequest{
		Username: claims.Username,
	})
	if err != nil {
		handleChatGrpcError(w, err)
		return
	}

	contacts := contactsToNormal(grpcResp.Contacts)

	resp := ServerResponse{
		Description: "OK",
		Data:        contacts,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func (h *ChatHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	type Response struct {
		ChatID     uint64 `json:"chat_id"`
		Avatar     string `json:"avatar"`
		PublicName string `json:"public_name"`
	}

	var user Username

	if err := DecodeData(w, r.Body, &user); err != nil {
		return
	}

	if user.Username == "" || claims.Username == user.Username {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	grpcResp, err := h.ChatService.CreateContact(ctx, &gen.CreateContactRequest{
		Username: claims.Username,
		TargetUsername: user.Username,
	})
	if err != nil {
		handleChatGrpcError(w, err)
		return
	}

	postResp := Response{
		ChatID:     grpcResp.ChatID,
		Avatar:     grpcResp.Avatar,
		PublicName: grpcResp.PublicName,
	}

	resp := ServerResponse{
		Description: "Created",
		Data:        postResp,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusCreated)
}

// GET /api/v1/chat?id=[id]
func (h *ChatHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	strID := r.URL.Query().Get("id")
	ID, err := strconv.Atoi(strID)
	if err != nil {
		HttpErrorToJson(w, "id must be an integer", http.StatusBadRequest)
		return
	}
	if ID <= 0 {
		HttpErrorToJson(w, "id must be greater than zero", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	grpcResp, err := h.ChatService.GetChat(ctx, &gen.GetChatRequest{
		ChatID: uint64(ID),
	})
	if err != nil {
		handleChatGrpcError(w, err)
		return
	}

	chat := domain.Chat{
		ChatID:       uint(grpcResp.ChatID),
		Username:     grpcResp.Username,
		Avatar:       grpcResp.Avatar,
		PublicName:   grpcResp.PublicName,
		MessageCount: uint(grpcResp.MessageCount),
		Messages:     messagesToNormal(grpcResp.Messages.Messages),
	}

	resp := ServerResponse{
		Description: "OK",
		Data:        chat,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

type MessageHandler func(ctx context.Context, conn *websocket.Conn, msg CommonWebsocket, claims *auth.Claims, hub *chatWebsocket.Hub) error

var handlers = map[string]MessageHandler{
    "message":  handleMessage,
    "mark_read": handleMarkRead,
}

type CommonWebsocket struct {
	Description string `json:"description"`
	Message string `json:"message"`
	ChatID int `json:"chat_id"`
	MessageID int `json:"message_id"`
	Username string `json:"username"`
	TargetUsername string `json:"target_username"`
}

func (h *ChatWebsocketHandler) WebSocketUpgrader(w http.ResponseWriter, r *http.Request) {
	var msg CommonWebsocket

    upgrader := websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        HttpErrorToJson(w, fmt.Errorf("failed to upgrade to websockets: %v", err).Error(), http.StatusInternalServerError)
        return
    }
    defer conn.Close()

    claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    h.Hub.AddClient(claims.Username, conn)

    for {
        err := conn.ReadJSON(&msg)
        if err != nil {
            log.Println("Error reading message:", err)
            break
        }

        description := msg.Description
        if description == "" {
            if err := conn.WriteJSON("bad request"); err != nil {
                log.Println("Failed to write error response:", err)
            }
            continue
        }

        handler, exists := handlers[description]
        if !exists {
            if err := conn.WriteJSON("unknown message type"); err != nil {
                log.Println("Failed to write error response:", err)
            }
            continue
        }

		log.Println("about to process socket")
        if err := handler(ctx, conn, msg, claims, h.Hub); err != nil {
            log.Printf("Error handling message type '%s': %v", description, err)
            if err := conn.WriteJSON(fmt.Sprintf("error processing %s", description)); err != nil {
                log.Println("Failed to write error response:", err)
            }
        }
    }
}

func handleMessage(ctx context.Context, conn *websocket.Conn, msg CommonWebsocket, claims *auth.Claims, hub *chatWebsocket.Hub) error {
    message := domain.Message{
        Content: msg.Message,
        ChatID:  uint64(msg.ChatID),
        Sender:  claims.Username,
    }

	log.Printf("sending a message to chat: %d", msg.ChatID)

    hub.Send(ctx, message, msg.Username)
    return nil
}

func handleMarkRead(ctx context.Context, conn *websocket.Conn, msg CommonWebsocket, claims *auth.Claims, hub *chatWebsocket.Hub) error {
    hub.MarkRead(ctx, msg.MessageID, msg.ChatID, msg.TargetUsername, claims.Username)
    return nil
}

func contactsToNormal(grpcContacts []*gen.Contact) []domain.Contact {
	var normal []domain.Contact

	for i := range grpcContacts {
		contact := grpcContacts[i]
		normal = append(normal, domain.Contact{
			Username:   contact.Username,
			Avatar:     contact.Avatar,
			PublicName: contact.PublicUsername,
		})
	}

	return normal
}

func chatsToNormal(grpcChats []*gen.Chat) []domain.Chat {
	var normal []domain.Chat

	for i := range grpcChats {
		chat := grpcChats[i]
		normal = append(normal, domain.Chat{
			ChatID:       uint(chat.ChatID),
			Username:     chat.Username,
			Avatar:       chat.Avatar,
			PublicName:   chat.PublicName,
			MessageCount: uint(chat.MessageCount),
			Messages:     messagesToNormal(chat.Messages.Messages),
			LastMessage:  &messagesToNormal([]*gen.Message{chat.LastMessage})[0],
		})
	}

	return normal
}

func messagesToNormal(grpcNormal []*gen.Message) []domain.Message {
	var normal []domain.Message

	for i := range grpcNormal {
		message := grpcNormal[i]
		normal = append(normal, domain.Message{
			MessageID: uint(message.MessageID),
			Content:   message.Content,
			Timestamp: message.Timestamp.AsTime(),
			IsRead:    message.IsRead,
			Sender:    message.Sender,
			Recipient: message.Recipient,
			ChatID:    message.ChatID,
		})
	}

	return normal
}
