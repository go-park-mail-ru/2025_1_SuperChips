package rest

import (
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pin"
	"github.com/go-park-mail-ru/2025_1_SuperChips/user"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
)

type AppHandler struct {
	Config      configs.Config
	UserService user.UserService
	PinService  pin.PinService
	JWTManager  rest.JWTManager
}

var ErrBadRequest = fmt.Errorf("bad request")

// HealthCheckHandler godoc
// @Summary Check server status
// @Description Returns server status
// @Produce json
// @Success 200 string serverResponse.Description
// @Router /health [get]
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := ServerResponse{
		Description: "server is up",
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}

