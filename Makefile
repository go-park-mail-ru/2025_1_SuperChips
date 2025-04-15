GO := go
MOCKGEN=mockgen
NAME := main.exe
COVERAGE_FILE=coverage.out
MOCK_DST=./mocks
TESTED_DIRS := ./internal/rest/... ./domain/... ./internal/repository/...

build : cmd/main.go
	$(GO) build -o $(NAME) $<

.PHONY : build test mocks

mocks:
	@mkdir -p $(MOCK_DST)/pin $(MOCK_DST)/user
	$(MOCKGEN) -source=./pin/service.go -destination=$(MOCK_DST)/pin/service.go
	$(MOCKGEN) -source=./user/service.go -destination=$(MOCK_DST)/user/service.go
	$(MOCKGEN) -source=./profile/service.go -destination=$(MOCK_DST)/profile/service.go
	$(MOCKGEN) -source=./internal/rest/profile.go -destination=$(MOCK_DST)/rest/profile.go


test: mocks
	go test $(TESTED_DIRS) -coverprofile=$(COVERAGE_FILE)
	
cover: test
	cat $(COVERAGE_FILE) | grep -v '_mock.go'
	go tool cover -html=$(COVERAGE_FILE)