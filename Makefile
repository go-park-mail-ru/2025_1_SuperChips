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
	$(MOCKGEN) -source=./pin/service.go -destination=$(MOCK_DST)/pin/repository/repository.go
	$(MOCKGEN) -source=./auth_service/auth/service.go -destination=$(MOCK_DST)/user/repository/repository.go
	$(MOCKGEN) -source=./profile/service.go -destination=$(MOCK_DST)/profile/repository/repository.go
	$(MOCKGEN) -source=./internal/rest/profile.go -destination=$(MOCK_DST)/profile/service/service.go
	$(MOCKGEN) -source=./board/service.go -destination=$(MOCK_DST)/board/repository/repository.go
	$(MOCKGEN) -source=./internal/rest/auth.go -destination=$(MOCK_DST)/user/service/service.go
	$(MOCKGEN) -source=./like/service.go -destination=$(MOCK_DST)/like/repository/repository.go
	$(MOCKGEN) -source=./internal/rest/like.go -destination=$(MOCK_DST)/like/service/service.go
	$(MOCKGEN) -source=./internal/rest/board.go -destination=$(MOCK_DST)/board/service/service.go


test: mocks
	go test $(TESTED_DIRS) -coverprofile=$(COVERAGE_FILE)
	
cover: mocks test
	cat $(COVERAGE_FILE) | grep -v 'mock_' | grep -v 'docs' | grep -v 'test_utils' > cover.out
	go tool cover -func=cover.out