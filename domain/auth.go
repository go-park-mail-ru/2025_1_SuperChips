package domain

//easyjson:json
type LoginData struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

//easyjson:json
type ExternalData struct {
	AccessToken string `json:"access_token,omitempty"`
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
}

//easyjson:json
type VKUser struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}

//easyjson:json
type CSRFResponse struct {
	CSRFToken string `json:"csrf_token"`
}

//easyjson:json
type RegisterData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

//easyjson:json
type VKUserTop struct {
	User VKUser `json:"user"`
}

