package test_utils

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
)

var TestConfig = configs.Config{
	ImageBaseDir:   "test_images", // Путь к папке с тестовыми изображениями
	IpAddress:      "localhost",
	Port:           "8080",
	PageSize:       20,
	Environment:    "test",
	JWTSecret:      []byte("j8reh9egjiofdhopsef"),
	ExpirationTime: time.Minute * 15,
}

var Host = "http://" + TestConfig.IpAddress + ":" + TestConfig.Port

func Marshal(body any) string {
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func GetBodyJson(rr *httptest.ResponseRecorder) string {
	body, _ := io.ReadAll(rr.Result().Body)
	return strings.Trim(string(body), " \n")
}

func PrintDifference(t *testing.T, name string, got any, exp any) {
	t.Errorf("Wrong: %+v", name)
	t.Errorf("--> got     : %+v", got)
	t.Errorf("--> expected: %+v", exp)
}
