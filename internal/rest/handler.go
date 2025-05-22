package rest

import (
	"fmt"
	"net/http"

)


var ErrBadRequest = fmt.Errorf("bad request")

// HealthCheckHandler godoc
//	@Summary		Check server status
//	@Description	Returns server status
//	@Produce		json
//	@Success		200	string	serverResponse.Description
//	@Router			/health [get]
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := ServerResponse{
		Description: "server is up",
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}
