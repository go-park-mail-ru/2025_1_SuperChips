package rest_test

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	mock_rest "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/rest"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type TestCase struct {
	Name         string
	Method       string
	URL          string
	Token        string
	Body         string
	ExpectedCode int
	ExpectedBody string
}

var conf = configs.Config{
	Port:           ":8080",
	JWTSecret:      []byte("secret"),
	ExpirationTime: 15 * time.Minute,
	CookieSecure:   false,
	Environment:    "test",
	IpAddress:      "localhost",
	ImageBaseDir:   "img",
	StaticBaseDir:  "static",
	AvatarDir:      "avatars",
	BaseUrl:        "http://localhost:8080/",
	PageSize:       10,
	AllowedOrigins: []string{"http://localhost:8080"},
}

func generateJWTToken(secret string) (string, error) {
	claims := auth.Claims{
		UserID: 1,
		Email:  "email@email.ru",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    "flow",
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestCurrentUserProfileHandler_GET(t *testing.T) {
    base := "/profile"

    validToken, err := generateJWTToken(string(conf.JWTSecret))
    if err != nil {
        t.Fatalf("failed to generate JWT token: %v", err)
    }

    testCases := []TestCase{
        {
            Name:         "Valid request",
            Method:       "GET",
            URL:          base,
            Token:        validToken,
            ExpectedCode: 200,
            ExpectedBody: `{"data":{"username":"JohnDoe","email":"","birthday":"2000-01-01T00:00:00Z"}}`,
        },
        {
            Name:         "Unauthorized request",
            Method:       "GET",
            URL:          base,
            ExpectedCode: 401,
            ExpectedBody: `{"description":"Unauthorized"}`,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.Name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockService := mock_rest.NewMockProfileService(ctrl)

            if tc.Name == "Valid request" {
                mockService.EXPECT().GetUserPublicInfoByEmail("email@email.ru").Return(domain.User{
                    Id:       1,
                    Birthday: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
                    Username: "JohnDoe",
                }, nil)
            }

            req := httptest.NewRequest(tc.Method, tc.URL, nil)
            if tc.Token != "" {
                req.AddCookie(&http.Cookie{
                    Name:  auth.AuthToken,
                    Value: tc.Token,
                })
            }

            handler := rest.ProfileHandler{
                ProfileService: mockService,
                JwtManager:     *auth.NewJWTManager(conf),
                AvatarFolder:   conf.AvatarDir,
                BaseUrl:        conf.BaseUrl,
                StaticFolder:   conf.StaticBaseDir,
                ExpirationTime: conf.ExpirationTime,
                CookieSecure:   conf.CookieSecure,
            }

            rr := httptest.NewRecorder()
            http.HandlerFunc(handler.CurrentUserProfileHandler).ServeHTTP(rr, req)

            if rr.Code != tc.ExpectedCode {
                t.Errorf("expected code %d, got %d", tc.ExpectedCode, rr.Code)
            }
            if strings.TrimSpace(rr.Body.String()) != tc.ExpectedBody {
                t.Errorf("expected body %s, got %s", tc.ExpectedBody, rr.Body.String())
            }
        })
    }
}

func TestCurrentUserProfileHandler_PATCH(t *testing.T) {
    base := "/profile"

    validToken, err := generateJWTToken(string(conf.JWTSecret))
    if err != nil {
        t.Fatalf("failed to generate JWT token: %v", err)
    }

    testCases := []TestCase{
        {
            Name:         "Patch profile",
            Method:       "PATCH",
            Body:         `{"public_name":"idk","birthday":"2000-02-01T00:00:00Z","about":"idk","email":"verynice@mail.ru"}`,
            URL:          base,
            Token:        validToken,
            ExpectedCode: 200,
            ExpectedBody: `{"description":"OK"}`,
        },
        {
            Name:         "bad request body",
            Method:       "PATCH",
            Body:         `{asfsafasfsafsa`,
            URL:          base,
            Token:        validToken,
            ExpectedCode: 400,
            ExpectedBody: `{"description":"Bad Request"}`,
        },
        {
            Name:         "patch validation error",
            Method:       "PATCH",
            Body:         `{"email":"invalidemail","birthday":"2000-02-01T00:00:00Z","about":"idk","public_name":"idk"}`,
            URL:          base,
            Token:        validToken,
            ExpectedCode: 400,
            ExpectedBody: `{"description":"validation failed"}`,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.Name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockService := mock_rest.NewMockProfileService(ctrl)

            switch tc.Name {
            case "Patch profile":
                mockService.EXPECT().GetUserPublicInfoByEmail("email@email.ru").Return(domain.User{
                    Id:       1,
                    Birthday: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
                    Username: "JohnDoe",
                }, nil)
                mockService.EXPECT().UpdateUserData(gomock.Any(), "email@email.ru").Return(nil)
            case "patch validation error":
                mockService.EXPECT().GetUserPublicInfoByEmail("email@email.ru").Return(domain.User{
                    Id:       1,
                    Birthday: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
                    Username: "JohnDoe",
                }, nil)
            }

            req := httptest.NewRequest(tc.Method, tc.URL, strings.NewReader(tc.Body))
            if tc.Token != "" {
                req.AddCookie(&http.Cookie{
                    Name:  auth.AuthToken,
                    Value: tc.Token,
                })
            }

            handler := rest.ProfileHandler{
                ProfileService: mockService,
                JwtManager:     *auth.NewJWTManager(conf),
                AvatarFolder:   conf.AvatarDir,
                BaseUrl:        conf.BaseUrl,
                StaticFolder:   conf.StaticBaseDir,
                ExpirationTime: conf.ExpirationTime,
                CookieSecure:   conf.CookieSecure,
            }

            rr := httptest.NewRecorder()
            http.HandlerFunc(handler.PatchUserProfileHandler).ServeHTTP(rr, req)

            if rr.Code != tc.ExpectedCode {
                t.Errorf("expected code %d, got %d", tc.ExpectedCode, rr.Code)
            }
            if strings.TrimSpace(rr.Body.String()) != tc.ExpectedBody {
                t.Errorf("expected body %s, got %s", tc.ExpectedBody, rr.Body.String())
            }
        })
    }
}

func TestPublicProfileHandler(t *testing.T) {
	validToken, err := generateJWTToken(string(conf.JWTSecret))
	if err != nil {
		t.Fatalf("failed to generate JWT token: %v", err)
	}

	testCases := []TestCase{
		{
			Name:         "Valid public profile",
			Method:       "GET",
			URL:          "/profile/johndoe",
			Token:        validToken,
			ExpectedCode: 200,
			ExpectedBody: `{"data":{"username":"johndoe","email":"","birthday":"0001-01-01T00:00:00Z","about":"Developer","public_name":"John Doe"}`,
		},
		{
			Name:         "Non-existent user",
			Method:       "GET",
			URL:          "/profile/unknown",
			Token:        validToken,
			ExpectedCode: 404,
			ExpectedBody: `{"description":"user not found"}`,
		},
		{
			Name:         "Invalid method",
			Method:       "POST",
			URL:          "/profile/johndoe",
			Token:        validToken,
			ExpectedCode: 405,
			ExpectedBody: `Method Not Allowed`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := mock_rest.NewMockProfileService(ctrl)

			mux := http.NewServeMux()
			handler := &rest.ProfileHandler{
				ProfileService: mockService,
				JwtManager:     *auth.NewJWTManager(conf),
			}
			mux.HandleFunc("GET /profile/{username}", handler.PublicProfileHandler)

			req := httptest.NewRequest(tc.Method, tc.URL, nil)
			if tc.Token != "" {
				req.AddCookie(&http.Cookie{
					Name:  auth.AuthToken,
					Value: tc.Token,
				})
			}

			switch tc.Name {
			case "Valid public profile":
				mockService.EXPECT().GetUserPublicInfoByUsername("johndoe").Return(domain.User{
					Username:   "johndoe",
					PublicName: "John Doe",
					About:      "Developer",
				}, nil)
			case "Non-existent user":
				mockService.EXPECT().GetUserPublicInfoByUsername("unknown").Return(domain.User{}, domain.ErrUserNotFound)
			}

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tc.ExpectedCode {
				t.Errorf("expected code %d, got %d", tc.ExpectedCode, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tc.ExpectedBody) {
				t.Errorf("expected body to contain %s, got %s", tc.ExpectedBody, rr.Body.String())
			}
		})
	}
}

func TestUserAvatarHandler(t *testing.T) {
	base := "/profile/avatar"

	validToken, err := generateJWTToken(string(conf.JWTSecret))
	if err != nil {
		t.Fatalf("failed to generate JWT token: %v", err)
	}

	testCases := []TestCase{
		{
			Name:         "Valid avatar upload",
			Method:       "POST",
			URL:          base,
			Token:        validToken,
			ExpectedCode: 201,
			ExpectedBody: `{"description":"Created"}`,
		},
		{
			Name:         "File too large",
			Method:       "POST",
			URL:          base,
			Token:        validToken,
			ExpectedCode: 413,
			ExpectedBody: `{"description":"image is too large"}`,
		},
		{
			Name:         "Invalid content type",
			Method:       "POST",
			URL:          base,
			Token:        validToken,
			ExpectedCode: 415,
			ExpectedBody: `{"description":"unsupported file format"}`,
		},
		{
			Name:         "No file uploaded",
			Method:       "POST",
			URL:          base,
			Token:        validToken,
			Body:         "",
			ExpectedCode: 400,
			ExpectedBody: `{"description":"Bad Request"}`,
		},
		{
			Name:         "unauthorized",
			Method:       "POST",
			URL:          base,
			Token:        "",
			Body:         "",
			ExpectedCode: 401,
			ExpectedBody: `{"description":"Unauthorized"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req := httptest.NewRequest(tc.Method, tc.URL, strings.NewReader(tc.Body))
			req.Header.Set("Content-Type", "multipart/form-data; boundary=TestBoundary")

			ctrl := gomock.NewController(t)
			mockService := mock_rest.NewMockProfileService(ctrl)

			handler := rest.ProfileHandler{
				ProfileService: mockService,
				JwtManager:     *auth.NewJWTManager(conf),
				AvatarFolder:   conf.AvatarDir,
				StaticFolder:   conf.StaticBaseDir,
				BaseUrl:        conf.BaseUrl,
			}

			switch tc.Name {
			case "Valid avatar upload":
				body, contentType := createMultipartFormData(t, "test.png", "image/png", 1024)
				req = httptest.NewRequest(tc.Method, tc.URL, body)
				req.Header.Set("Content-Type", contentType)
				mockService.EXPECT().SaveUserAvatar(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			case "File too large":
				body, contentType := createMultipartFormData(t, "large.jpg", "image/jpeg", 4*1024*1024)
				req = httptest.NewRequest(tc.Method, tc.URL, body)
				req.Header.Set("Content-Type", contentType)
			case "Invalid content type":
				body, contentType := createMultipartFormData(t, "test.txt", "text/plain")
				req = httptest.NewRequest(tc.Method, tc.URL, body)
				req.Header.Set("Content-Type", contentType)
			case "No file uploaded":
				req = httptest.NewRequest(tc.Method, tc.URL, nil)
				req.Header.Set("Content-Type", "multipart/form-data; boundary=TestBoundary")
			}

			if tc.Token != "" {
				req.AddCookie(&http.Cookie{
					Name:     auth.AuthToken,
					Value:    tc.Token,
					Path:     "/",
					Expires:  time.Now().Add(15 * time.Minute),
					Secure:   conf.CookieSecure,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
			}

			rr := httptest.NewRecorder()
			http.HandlerFunc(handler.UserAvatarHandler).ServeHTTP(rr, req)

			if rr.Code != tc.ExpectedCode {
				t.Errorf("expected code %d, got %d", tc.ExpectedCode, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tc.ExpectedBody) {
				t.Errorf("expected body to contain %s, got %s", tc.ExpectedBody, rr.Body.String())
			}
		})
	}
}

func TestChangeUserPasswordHandler(t *testing.T) {
	base := "/profile/password"

	validToken, err := generateJWTToken(string(conf.JWTSecret))
	if err != nil {
		t.Fatalf("failed to generate JWT token: %v", err)
	}

	testCases := []TestCase{
		{
			Name:         "Valid password change",
			Method:       "POST",
			URL:          base,
			Token:        validToken,
			Body:         `{"old_password":"oldpass","new_password":"NewPass123!"}`,
			ExpectedCode: 200,
			ExpectedBody: `{"description":"OK"}`,
		},
		{
			Name:         "Incorrect old password",
			Method:       "POST",
			URL:          base,
			Token:        validToken,
			Body:         `{"old_password":"wrongpass","new_password":"NewPass123!"}`,
			ExpectedCode: 401,
			ExpectedBody: `{"description":"Unauthorized"}`,
		},
		{
			Name:         "Invalid new password",
			Method:       "POST",
			URL:          base,
			Token:        validToken,
			Body:         `{"old_password":"oldpass","new_password":""}`,
			ExpectedCode: 400,
			ExpectedBody: `{"description":"cannot use empty password"}`,
		},
		{
			Name:         "Missing fields",
			Method:       "POST",
			URL:          base,
			Token:        validToken,
			Body:         `{"old_password":"oldpass"}`,
			ExpectedCode: 400,
			ExpectedBody: `{"description":"cannot use empty password"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req := httptest.NewRequest(tc.Method, tc.URL, strings.NewReader(tc.Body))
			req.Header.Set("Content-Type", "application/json")

			if tc.Token != "" {
				req.AddCookie(&http.Cookie{
					Name:  auth.AuthToken,
					Value: tc.Token,
				})
			}

			ctrl := gomock.NewController(t)
			mockService := mock_rest.NewMockProfileService(ctrl)

			handler := rest.ProfileHandler{
				ProfileService: mockService,
				JwtManager:     *auth.NewJWTManager(conf),
				ExpirationTime: conf.ExpirationTime,
				CookieSecure:   conf.CookieSecure,
			}

			switch tc.Name {
			case "Valid password change":
				mockService.EXPECT().ChangeUserPassword("email@email.ru", "oldpass", "NewPass123!").Return(1, nil)
			case "Incorrect old password":
				mockService.EXPECT().ChangeUserPassword("email@email.ru", "wrongpass", "NewPass123!").Return(0, domain.ErrInvalidCredentials)
			case "Invalid new password":
				mockService.EXPECT().ChangeUserPassword("email@email.ru", "oldpass", "").Return(0, domain.ErrNoPassword)
			case "Missing fields":
				mockService.EXPECT().ChangeUserPassword(gomock.Any(), gomock.Any(), "").Return(0, domain.ErrNoPassword)
			}

			rr := httptest.NewRecorder()
			http.HandlerFunc(handler.ChangeUserPasswordHandler).ServeHTTP(rr, req)

			if rr.Code != tc.ExpectedCode {
				t.Errorf("expected code %d, got %d", tc.ExpectedCode, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tc.ExpectedBody) {
				t.Errorf("expected body to contain %s, got %s", tc.ExpectedBody, rr.Body.String())
			}
		})
	}
}

func createMultipartFormData(t *testing.T, filename, contentType string, size ...int) (*bytes.Buffer, string) {
	var fileSize int
	if len(size) > 0 {
		fileSize = size[0]
	} else {
		fileSize = 1024
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	err := writer.SetBoundary("TestBoundary")
	if err != nil {
		t.Fatalf("failed to set boundary: %v", err)
	}

	headers := make(textproto.MIMEHeader)
	headers.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, filename))
	headers.Set("Content-Type", contentType)

	part, err := writer.CreatePart(headers)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	data := make([]byte, fileSize)
	_, err = part.Write(data)
	if err != nil {
		t.Fatalf("failed to write data to part: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	return body, writer.FormDataContentType()
}
