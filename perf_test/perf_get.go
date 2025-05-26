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
	targetURL   = "http://localhost:8080/api/v1/flows?id=1215"
	rate        = 200
	duration    = 500 * time.Second
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
		fmt.Printf("Error creating login request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error during login: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Login request returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		os.Exit(1)
	}

	var tokens Tokens
	err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&tokens)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		os.Exit(1)
	}

	csrfToken := tokens.Data.CsrfToken
	var authToken string

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "auth_token" {
			authToken = cookie.Value
			break
		}
	}

	targets := fmt.Sprintf("GET %s\nX-CSRF-TOKEN: %s\nCookie: auth_token=%s; csrf_token=%s\n",
		targetURL, csrfToken, authToken, csrfToken)

	targetsFilePath := "vegeta_targets.txt"
	targetsFile, err := os.Create(targetsFilePath)
	if err != nil {
		fmt.Printf("Error creating targets file: %v\n", err)
		return
	}
	defer os.Remove(targetsFilePath)

	_, err = targetsFile.WriteString(targets)
	if err != nil {
		fmt.Printf("Error writing to targets file: %v\n", err)
		targetsFile.Close()
		return
	}
	targetsFile.Close()

	rateStr := strconv.Itoa(rate)
	durationStr := duration.String()

	cmd := exec.Command(vegetaBin, "attack",
		"-targets="+targetsFilePath,
		"-rate="+rateStr,
		"-duration="+durationStr,
		"-output="+resultsFile,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Starting Vegeta load test...\n")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running Vegeta: %v\n", err)
		return
	}

	fmt.Printf("Load test completed. Results saved to '%s'.\n", resultsFile)
	fmt.Printf("To view the report:\n%s report < %s\n", vegetaBin, resultsFile)
	fmt.Printf("To generate a plot:\n%s plot < %s > plot.html\n", vegetaBin, resultsFile)
}
