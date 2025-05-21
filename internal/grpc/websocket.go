package grpc

import (
	"context"
	"encoding/json"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	chatWebsocket "github.com/go-park-mail-ru/2025_1_SuperChips/internal/websocket"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/websocket"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcWebsocketHandler struct {
	gen.UnimplementedWebsocketServer
	hub *chatWebsocket.Hub
}

func genToDomainWebMessage(msg *gen.WebMessage) (domain.WebMessage, error) {
	var contentMap map[string]interface{}

	if msg.Content != nil {
		jsonData, err := protojson.Marshal(msg.Content)
		if err != nil {
			return domain.WebMessage{}, err
		}

		if err := json.Unmarshal(jsonData, &contentMap); err != nil {
			return domain.WebMessage{}, err
		}
	}

	return domain.WebMessage{
		Type:    msg.Type,
		Content: contentMap,
	}, nil
}

func NewGrpcWebsocketHandler(hub *chatWebsocket.Hub) *GrpcWebsocketHandler {
	return &GrpcWebsocketHandler{
		hub: hub,
	}
}

func (c *GrpcWebsocketHandler) SendWebMessage(ctx context.Context, in *gen.SendWebMessageRequest) (*emptypb.Empty, error) {
	webMsg, err := genToDomainWebMessage(in.WebMessage)
	if err != nil {
		return nil, err
	}
	
	if err := c.hub.SendNotification(ctx, webMsg); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}