GO := go
NAME := main.exe
COVERAGE_FILE=coverage.out
TESTED_DIRS := ./internal/auth/... ./internal/feed/... ./internal/handler/... ./internal/user/...

build : cmd/main.go
	$(GO) build -o $(NAME) $<

.PHONY : build

test:
	go test $(TESTED_DIRS) -coverprofile=$(COVERAGE_FILE)

cover: test
	cat $(COVERAGE_FILE) | grep -v '_mock.go'
	go tool cover -html=$(COVERAGE_FILE)