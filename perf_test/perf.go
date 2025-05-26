package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	targetURL   = "http://localhost:8080/api/v1/flows"
	imagePath   = "skebob.jpg"
	rate        = 40
	duration    = 2560 * time.Second
	boundary    = "827d5621cefb4d10806f3bf336e557e5"
	vegetaBin   = "vegeta"
	resultsFile = "vegeta_results.bin"
)

type Data struct {
	CsrfToken string `json:"csrf_token"`
}

type Tokens struct {
	Data Data `json:"data"`
}

func main() {
	loginURL := "http://localhost:8080/api/v1/auth/login"
	email := "load@test.ru"
	password := "11111111"

	requestBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{"email": "%s", "password": "%s"}`, email, password)))

	client := &http.Client{}

	req, err := http.NewRequest("POST", loginURL, requestBody)
	if err != nil {
		fmt.Printf("Ошибка при авторизации: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Ошибка при авторизации: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Ошибка: Запрос к '%s' вернул статус %d\n", loginURL, resp.StatusCode)
		os.Exit(1)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Ошибка при чтении тела ответа: %v\n", err)
		os.Exit(1)
	}

	var tokens Tokens
	err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&tokens)
	if err != nil {
		fmt.Printf("Ошибка при декодировании JSON: %v\n", err)
		os.Exit(1)
	}

	csrfToken := tokens.Data.CsrfToken

	var auth_token string

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "auth_token" {
			auth_token = cookie.Value
			break
		}
	}

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		fmt.Printf("Ошибка: Файл с картинкой '%s' не найден.\n", imagePath)
		return
	}

	body := new(bytes.Buffer)

	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=\"image\"; filename=\"%s\"\r\n", imagePath))
	body.WriteString("Content-Type: image/jpeg\r\n")
	body.WriteString("\r\n")

	imageFile, err := os.Open(imagePath)
	if err != nil {
		fmt.Printf("Ошибка при открытии файла с картинкой: %v\n", err)
		return
	}
	defer imageFile.Close()

	_, err = io.Copy(body, imageFile)
	if err != nil {
		fmt.Printf("Ошибка при чтении файла с картинкой: %v\n", err)
		return
	}
	body.WriteString("\r\n")
	body.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	bodyFilePath := "body_for_vegeta.txt"
	bodyFile, err := os.Create(bodyFilePath)
	if err != nil {
		fmt.Printf("Ошибка при создании временного файла %s: %v\n", bodyFilePath, err)
		return
	}
	_, err = bodyFile.Write(body.Bytes())
	if err != nil {
		fmt.Printf("Ошибка при записи тела запроса во временный файл: %v\n", err)
		bodyFile.Close()
		os.Remove(bodyFilePath)
		return
	}
	bodyFile.Close()
	defer os.Remove(bodyFilePath)

	targets := fmt.Sprintf("POST %s\nContent-Type: multipart/form-data; boundary=%s\nX-CSRF-TOKEN: %s\nCookie: auth_token=%s; csrf_token=%s\n@%s",
		targetURL, boundary, csrfToken, auth_token, csrfToken, bodyFilePath)

	targetsFilePath := "targets_for_vegeta.txt"
	targetsFile, err := os.Create(targetsFilePath)
	if err != nil {
		fmt.Printf("Ошибка при создании временного файла %s: %v\n", targetsFilePath, err)
		return
	}
	_, err = targetsFile.WriteString(targets)
	if err != nil {
		fmt.Printf("Ошибка при записи targets во временный файл: %v\n", err)
		targetsFile.Close()
		os.Remove(targetsFilePath)
		return
	}
	targetsFile.Close()
	defer os.Remove(targetsFilePath)

	rateStr := strconv.Itoa(rate)
	durationStr := duration.String()
	print(durationStr)

	cmd := exec.Command(vegetaBin, "attack",
		"-targets="+targetsFilePath,
		"-rate="+rateStr,
		"-duration="+durationStr,
		"-output="+resultsFile,
		"-insecure",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Запуск нагрузочного тестирования Vegeta...\n")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Ошибка при выполнении Vegeta: %v\n", err)
		return
	}

	fmt.Printf("Нагрузочное тестирование завершено. Результаты сохранены в '%s'.\n", resultsFile)
	fmt.Printf("Для просмотра отчета выполните:\n%s report < %s\n", vegetaBin, resultsFile)
	fmt.Printf("Для создания графика выполните:\n%s plot < %s > plot.html\n", vegetaBin, resultsFile)
}
