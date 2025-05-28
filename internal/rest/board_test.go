package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	mocks "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/board/service"
	"go.uber.org/mock/gomock"
)

type serverResponse struct {
	Description string          `json:"description"`
	Data        json.RawMessage `json:"data,omitempty"`
}

func newTestRequest(method, target string, body []byte, pathValues map[string]string) *http.Request {
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	return req
}

func TestCreateBoard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardService := mocks.NewMockBoardService(ctrl)
	handler := &BoardHandler{
		BoardService:    mockBoardService,
		ContextDeadline: 2 * time.Second,
	}

	claims := &auth.Claims{UserID: 111}

	boardPayload := domain.Board{
		Name:      "My Board",
		IsPrivate: false,
	}
	payloadBytes, err := json.Marshal(boardPayload)
	if err != nil {
		t.Fatal(err)
	}

	pathValues := map[string]string{"username": "testuser"}
	req := newTestRequest(http.MethodPost, "/api/v1/users/{username}/boards", payloadBytes, pathValues)
	req.SetPathValue("username", "testuser")
	ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	expectedBoard := boardPayload
	expectedBoard.AuthorID = claims.UserID
	mockBoardService.EXPECT().
		CreateBoard(gomock.Any(), gomock.AssignableToTypeOf(expectedBoard), "testuser", claims.UserID).
		Return(100, nil)

	rr := httptest.NewRecorder()
	handler.CreateBoard(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, rr.Code)
	}

	var resp serverResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if resp.Description != "OK" {
		t.Errorf("expected description OK; got %s", resp.Description)
	}

	var data struct {
		BoardID int `json:"board_id"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed decoding data: %v", err)
	}
	if data.BoardID != 100 {
		t.Errorf("expected board_id 100; got %d", data.BoardID)
	}
}

func TestDeleteBoard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardService := mocks.NewMockBoardService(ctrl)
	handler := &BoardHandler{
		BoardService:    mockBoardService,
		ContextDeadline: 2 * time.Second,
	}

	claims := &auth.Claims{UserID: 222}

	pathValues := map[string]string{"board_id": "123"}
	req := newTestRequest(http.MethodDelete, "/api/v1/boards/{board_id}", nil, pathValues)
	req.SetPathValue("board_id", "123")
	req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, claims))

	mockBoardService.EXPECT().
		DeleteBoard(gomock.Any(), 123, claims.UserID).
		Return(nil)

	rr := httptest.NewRecorder()
	handler.DeleteBoard(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, rr.Code)
	}
}

func TestAddToBoard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardService := mocks.NewMockBoardService(ctrl)
	handler := &BoardHandler{
		BoardService:    mockBoardService,
		ContextDeadline: 2 * time.Second,
	}

	claims := &auth.Claims{UserID: 333}
	pathValues := map[string]string{"id": "200"}
	boardReq := domain.BoardRequest{FlowID: 555}
	payloadBytes, err := json.Marshal(boardReq)
	if err != nil {
		t.Fatal(err)
	}
	req := newTestRequest(http.MethodPost, "/api/v1/boards/{board_id}/flows", payloadBytes, pathValues)
	req.SetPathValue("id", "200")
	req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, claims))

	mockBoardService.EXPECT().
		AddToBoard(gomock.Any(), 200, claims.UserID, boardReq.FlowID).
		Return(nil)

	rr := httptest.NewRecorder()
	handler.AddToBoard(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, rr.Code)
	}
}

func TestDeleteFromBoard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardService := mocks.NewMockBoardService(ctrl)
	handler := &BoardHandler{
		BoardService:    mockBoardService,
		ContextDeadline: 2 * time.Second,
	}

	claims := &auth.Claims{UserID: 444}
	pathValues := map[string]string{"board_id": "300", "id": "600"}
	req := newTestRequest(http.MethodDelete, "/api/v1/boards/{board_id}/flows/{id}", nil, pathValues)
	req.SetPathValue("board_id", "300")
	req.SetPathValue("id", "600")
	req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, claims))

	mockBoardService.EXPECT().
		DeleteFromBoard(gomock.Any(), 300, claims.UserID, 600).
		Return(nil)

	rr := httptest.NewRecorder()
	handler.DeleteFromBoard(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, rr.Code)
	}
}

func TestUpdateBoard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardService := mocks.NewMockBoardService(ctrl)
	handler := &BoardHandler{
		BoardService:    mockBoardService,
		ContextDeadline: 2 * time.Second,
	}

	claims := &auth.Claims{UserID: 555}
	pathValues := map[string]string{"board_id": "400"}
	updatePayload := map[string]interface{}{
		"name":       "Updated Board",
		"is_private": true,
	}
	payloadBytes, err := json.Marshal(updatePayload)
	if err != nil {
		t.Fatal(err)
	}
	req := newTestRequest(http.MethodPut, "/api/v1/boards/{board_id}", payloadBytes, pathValues)
	req.SetPathValue("board_id", "400")
	req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, claims))

	mockBoardService.EXPECT().
		UpdateBoard(gomock.Any(), 400, claims.UserID, "Updated Board", true).
		Return(nil)

	rr := httptest.NewRecorder()
	handler.UpdateBoard(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, rr.Code)
	}
}

func TestGetBoard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardService := mocks.NewMockBoardService(ctrl)
	handler := &BoardHandler{
		BoardService:    mockBoardService,
		ContextDeadline: 2 * time.Second,
	}

	claims := &auth.Claims{UserID: 666}
	pathValues := map[string]string{"board_id": "500"}
	req := newTestRequest(http.MethodGet, "/api/v1/boards/{board_id}", nil, pathValues)
	req.SetPathValue("board_id", "500")
	req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, claims))

	dummyBoard := domain.Board{
		ID:             500,
		AuthorID:       666,
		Name:           "Dummy Board",
		IsPrivate:      false,
		FlowCount:      10,
		AuthorUsername: "dummyuser",
	}
	mockBoardService.EXPECT().
		GetBoard(gomock.Any(), 500, claims.UserID, true).
		Return(dummyBoard, nil)

	rr := httptest.NewRecorder()
	handler.GetBoard(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, rr.Code)
	}

	var resp serverResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	var boardResp domain.Board
	if err := json.Unmarshal(resp.Data, &boardResp); err != nil {
		t.Fatalf("failed decoding board data: %v", err)
	}
	if boardResp.ID != dummyBoard.ID || boardResp.Name != dummyBoard.Name {
		t.Errorf("expected board %+v; got %+v", dummyBoard, boardResp)
	}
}

func TestGetUserPublic_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardService := mocks.NewMockBoardService(ctrl)
	handler := &BoardHandler{
		BoardService:    mockBoardService,
		ContextDeadline: 2 * time.Second,
	}

	pathValues := map[string]string{"username": "publicuser"}
	req := newTestRequest(http.MethodGet, "/api/v1/user/{username}/boards", nil, pathValues)
	req.SetPathValue("username", "publicuser")

	dummyBoards := []domain.Board{
		{ID: 1, Name: "Board One"},
		{ID: 2, Name: "Board Two"},
	}

	mockBoardService.EXPECT().
		GetUserPublicBoards(gomock.Any(), "publicuser").
		Return(dummyBoards, nil)

	rr := httptest.NewRecorder()
	handler.GetUserPublic(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, rr.Code)
	}
}

func TestGetUserAllBoards_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardService := mocks.NewMockBoardService(ctrl)
	handler := &BoardHandler{
		BoardService:    mockBoardService,
		ContextDeadline: 2 * time.Second,
	}

	claims := &auth.Claims{UserID: 777}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/boards", nil)
	req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, claims))
	rr := httptest.NewRecorder()

	dummyBoards := []domain.Board{
		{ID: 10, Name: "Private Board"},
		{ID: 20, Name: "Public Board"},
	}
	mockBoardService.EXPECT().
		GetUserAllBoards(gomock.Any(), claims.UserID).
		Return(dummyBoards, nil)

	handler.GetUserAllBoards(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, rr.Code)
	}
}

func TestGetBoardFlows_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardService := mocks.NewMockBoardService(ctrl)
	handler := &BoardHandler{
		BoardService:    mockBoardService,
		ContextDeadline: 2 * time.Second,
	}

	claims := &auth.Claims{UserID: 888}
	pathValues := map[string]string{"board_id": "700"}
	req := newTestRequest(http.MethodGet, "/api/v1/boards/{board_id}/flows?page=2&size=10", nil, pathValues)
	req.SetPathValue("board_id", "700")
	req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, claims))

	page, size := 2, 10
	dummyFlows := []domain.PinData{
		{FlowID: 1},
		{FlowID: 2},
	}
	mockBoardService.EXPECT().
		GetBoardFlow(gomock.Any(), 700, claims.UserID, page, size, true).
		Return(dummyFlows, nil)

	rr := httptest.NewRecorder()
	handler.GetBoardFlows(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, rr.Code)
	}
}
