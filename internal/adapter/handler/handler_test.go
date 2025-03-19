// package handler_test

// import (
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
// 	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/auth"
// 	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/feed"
// 	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/handler"
// 	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
// )

// var cfg configs.Config

// func init() {
// 	cfg = configs.Config{
// 		ImageBaseDir: "test_images", // Путь к папке с тестовыми изображениями
// 		IpAddress:    "localhost",
// 		Port:         "8080",
// 		PageSize: 20,
// 		Environment: "test",
// 		JWTSecret: []byte("j8reh9egjiofdhopsef"),
// 		ExpirationTime: time.Minute * 15,
// 	}
// }

// func printDifference(t *testing.T, num int, name string, got any, exp any) {
// 	t.Errorf("[%d] wrong %v", num, name)
// 	t.Errorf("--> got     : %+v", got)
// 	t.Errorf("--> expected: %+v", exp)
// }

// func TestHealthCheckHandler(t *testing.T) {
// 	type TestCase struct {
// 		Response   string
// 		StatusCode int
// 	}

// 	cases := []TestCase{
// 		{
// 			Response:   `{"description":"server is up"}`,
// 			StatusCode: http.StatusOK,
// 		},
// 	}

// 	for num, c := range cases {
// 		url := "http://localhost:8080/health"
// 		req := httptest.NewRequest("GET", url, nil)
// 		w := httptest.NewRecorder()

// 		app := handler.AppHandler{}
// 		app.HealthCheckHandler(w, req)

// 		if w.Code != c.StatusCode {
// 			printDifference(t, num, "StatusCode", w.Code, c.StatusCode)
// 		}

// 		resp := w.Result()
// 		body, _ := io.ReadAll(resp.Body)
// 		bodyStr := strings.Trim(string(body), " \n")

// 		if bodyStr != c.Response {
// 			printDifference(t, num, "Response", bodyStr, c.Response)
// 		}
// 	}
// }

// func TestFeedHandler(t *testing.T) {
// 	type TestCase struct {
// 		Page       string
// 		PageSize   int
// 		Response   string
// 		StatusCode int
// 	}

// 	cases := []TestCase{
// 		// Сценарий: некорректный page (page < 1). В рамках реализации исправляется на корректный.
// 		{
// 			Page:       "0",
// 			PageSize:   1,
// 			Response:   `{"data":[{"header":"Header 1","image":"http://localhost:8080/static/img/image1.png","author":"Author -1"}]}`,
// 			StatusCode: http.StatusOK,
// 		},
// 		// Сценарий: страница существует, но на ней нет данных.
// 		{
// 			Page:       "4",
// 			PageSize:   2,
// 			Response:   `{"description":"Not Found"}`,
// 			StatusCode: http.StatusNotFound,
// 		},
// 		// Сценарий: страница существует и на ней есть данные.
// 		{
// 			Page:       "1",
// 			PageSize:   2,
// 			Response:   `{"data":[{"header":"Header 1","image":"http://localhost:8080/static/img/image1.png","author":"Author -1"},{"header":"Header 2","image":"http://localhost:8080/static/img/image2.png","author":"Author -2"}]}`,
// 			StatusCode: http.StatusOK,
// 		},
// 	}

// 	mockNewPinStorage := func() feed.PinStorage {
// 		p := feed.NewPinSliceStorage(cfg)
// 		p.Pins = append(p.Pins, feed.PinData{
// 			Header: fmt.Sprintf("Header %d", 1),
// 			Image:  fmt.Sprintf("http://localhost:8080/static/img/%s", "image1.png"),
// 			Author: fmt.Sprintf("Author %d", -1),
// 		})
// 		p.Pins = append(p.Pins, feed.PinData{
// 			Header: fmt.Sprintf("Header %d", 2),
// 			Image:  fmt.Sprintf("http://localhost:8080/static/img/%s", "image2.png"),
// 			Author: fmt.Sprintf("Author %d", -2),
// 		})
// 		return p
// 	}

// 	for num, c := range cases {
// 		url := fmt.Sprintf("http://localhost:8080/api/v1/feed?page=%s", c.Page)
// 		req := httptest.NewRequest("GET", url, nil)
// 		w := httptest.NewRecorder()

// 		app := handler.AppHandler{}
// 		app.Config.PageSize = c.PageSize
// 		app.PinStorage = mockNewPinStorage()

// 		app.FeedHandler(w, req)

// 		if w.Code != c.StatusCode {
// 			printDifference(t, num, "StatusCode", w.Code, c.StatusCode)
// 		}

// 		resp := w.Result()
// 		body, _ := io.ReadAll(resp.Body)
// 		bodyStr := strings.Trim(string(body), " \n")

// 		if bodyStr != c.Response {
// 			printDifference(t, num, "Response", bodyStr, c.Response)
// 		}
// 	}
// }

// func TestLoginHandler(t *testing.T) {
// 	type TestCase struct {
// 		RequestBody string
// 		Response    string
// 		StatusCode  int
// 	}

// 	cases := []TestCase{
// 		// Сценарий: некорретное тело запроса.
// 		{
// 			RequestBody: `{bibi}`,
// 			Response:    `{"description":"Bad Request"}`,
// 			StatusCode:  http.StatusBadRequest,
// 		},
// 		// Сценарий: некорректные учётные данные/нет в БД.
// 		{
// 			RequestBody: `{"email": "void@example.com", "password": "void1"}`,
// 			Response:    `{"description":"invalid credentials"}`,
// 			StatusCode:  http.StatusUnauthorized,
// 		},
// 		// Сценарий: учётные данные есть в БД, но индекс в БД некорректный.
// 		{
// 			RequestBody: `{"email": "test3@example.com", "password": "pass3"}`,
// 			Response:    `{"description":"Internal Server Error"}`,
// 			StatusCode:  http.StatusInternalServerError,
// 		},
// 		// Сценарий: позитивный.
// 		{
// 			RequestBody: `{"email":"test1@example.com","password":"pass1"}`,
// 			Response:    `{"description":"OK"}`,
// 			StatusCode:  http.StatusOK,
// 		},
// 	}

// 	mockNewUserStorage := func() *user.MapUserStorage {
// 		strg := user.NewMapUserStorage()
// 		user1 := user.User{
// 			Username: "test1",
// 			Password: "pass1",
// 			Email:    "test1@example.com",
// 			Birthday: time.Date(2005, time.April, 4, 0, 0, 0, 0, time.UTC),
// 		}
// 		user2 := user.User{
// 			Username: "test2",
// 			Password: "pass2",
// 			Email:    "test2@example.com",
// 			Birthday: time.Date(2005, time.April, 4, 0, 0, 0, 0, time.UTC),
// 		}
// 		user_broken := user.User{
// 			Username: "test3",
// 			Password: "pass3",
// 			Email:    "test3@example.com",
// 			Birthday: time.Date(2005, time.April, 4, 0, 0, 0, 0, time.UTC),
// 		}
// 		if err := strg.AddUser(user1); err != nil {
// 			println(err)
// 		}
// 		if err := strg.AddUser(user2); err != nil {
// 			println(err)
// 		}

// 		if err := strg.AddUser(user_broken); err != nil {
// 			println(err)
// 		}
// 		strg.SetUserID(0, "test3@example.com")
// 		return strg
// 	}

// 	app := handler.AppHandler{}
// 	app.Config = cfg
// 	app.JWTManager = auth.NewJWTManager(cfg)
// 	app.UserStorage = mockNewUserStorage()

// 	for num, c := range cases {
// 		url := "http://localhost:8080//api/v1/auth/login"
// 		req := httptest.NewRequest("POST", url, strings.NewReader(c.RequestBody))
// 		w := httptest.NewRecorder()

// 		app.LoginHandler(w, req)

// 		if w.Code != c.StatusCode {
// 			printDifference(t, num, "StatusCode", w.Code, c.StatusCode)
// 		}

// 		resp := w.Result()
// 		body, _ := io.ReadAll(resp.Body)
// 		bodyStr := strings.Trim(string(body), " \n")

// 		if bodyStr != c.Response {
// 			printDifference(t, num, "Response", bodyStr, c.Response)
// 		}
// 	}
// }

// func TestRegistrationHandler(t *testing.T) {
// 	type TestCase struct {
// 		RequestBody string
// 		Response    string
// 		StatusCode  int
// 	}

// 	cases := []TestCase{
// 		// Сценарий: некорректное тело запроса.
// 		{
// 			RequestBody: `{bibi}`,
// 			Response:    `{"description":"Bad Request"}`,
// 			StatusCode:  http.StatusBadRequest,
// 		},
// 		// Сценарий: отсутствует имя пользователя.
// 		{
// 			RequestBody: `{"email": "void@example.com", "password": "void1"}`,
// 			Response:    `{"description":"Invalid username"}`,
// 			StatusCode:  http.StatusBadRequest,
// 		},
// 		// Сценарий: отсутствует дата рождения.
// 		{
// 			RequestBody: `{"email": "test1@example.com", "password": "pass1", "username": "test1"}`,
// 			Response:    `{"description":"Invalid birthday"}`,
// 			StatusCode:  http.StatusBadRequest,
// 		},
// 		// Сценарий: некорректная дата рождения.
// 		{
// 			RequestBody: `{"email": "test@example.com", "password": "password123", "birthday": "invalid-date"}`,
// 			Response:    `{"description":"Bad Request"}`,
// 			StatusCode:  http.StatusBadRequest,
// 		},
// 	}

// 	mockNewUserStorage := func() *user.MapUserStorage {
// 		strg := user.NewMapUserStorage()
// 		user1 := user.User{
// 			Username: "test1",
// 			Password: "pass1",
// 			Email:    "test1@example.com",
// 			Birthday: time.Date(2005, time.April, 4, 0, 0, 0, 0, time.UTC),
// 		}
// 		user2 := user.User{
// 			Username: "test2",
// 			Password: "pass2",
// 			Email:    "test2@example.com",
// 			Birthday: time.Date(2005, time.April, 4, 0, 0, 0, 0, time.UTC),
// 		}
// 		user_broken := user.User{
// 			Username: "test3",
// 			Password: "pass3",
// 			Email:    "test3@example.com",
// 			Birthday: time.Date(2005, time.April, 4, 0, 0, 0, 0, time.UTC),
// 		}
// 		strg.AddUser(user1)
// 		strg.AddUser(user2)
// 		strg.AddUser(user_broken)
// 		strg.SetUserID(0, "test3@example.com")
// 		return strg
// 	}

// 	app := handler.AppHandler{}
// 	app.UserStorage = mockNewUserStorage()

// 	for num, c := range cases {
// 		url := "http://localhost:8080/api/v1/auth/registration"
// 		req := httptest.NewRequest("POST", url, strings.NewReader(c.RequestBody))
// 		w := httptest.NewRecorder()

// 		app.RegistrationHandler(w, req)

// 		if w.Code != c.StatusCode {
// 			printDifference(t, num, "StatusCode", w.Code, c.StatusCode)
// 		}

// 		resp := w.Result()
// 		body, _ := io.ReadAll(resp.Body)
// 		bodyStr := strings.Trim(string(body), " \n")

// 		if bodyStr != c.Response {
// 			printDifference(t, num, "Response", bodyStr, c.Response)
// 		}
// 	}
// }
