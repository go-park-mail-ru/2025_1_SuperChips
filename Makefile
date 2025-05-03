GO := go
MOCKGEN=mockgen
NAME := main.exe
COVERAGE_FILE=coverage.out
MOCK_DST=./mocks
TESTED_DIRS := ./internal/rest/... ./domain/... ./internal/repository/...

.PHONY : test mocks

mocks:
	@mkdir -p $(MOCK_DST)/pin $(MOCK_DST)/user
	$(MOCKGEN) -source=./pin/service.go -destination=$(MOCK_DST)/pin/repository/repository.go
	$(MOCKGEN) -source=./auth/service.go -destination=$(MOCK_DST)/user/repository/repository.go
	$(MOCKGEN) -source=./profile/service.go -destination=$(MOCK_DST)/profile/repository/repository.go
	$(MOCKGEN) -source=./internal/rest/profile.go -destination=$(MOCK_DST)/profile/service/service.go
	$(MOCKGEN) -source=./board/service.go -destination=$(MOCK_DST)/board/repository/repository.go
	$(MOCKGEN) -source=./internal/grpc/auth.go -destination=$(MOCK_DST)/user/service/service.go
	$(MOCKGEN) -source=./like/service.go -destination=$(MOCK_DST)/like/repository/repository.go
	$(MOCKGEN) -source=./internal/rest/like.go -destination=$(MOCK_DST)/like/service/service.go
	$(MOCKGEN) -source=./internal/rest/board.go -destination=$(MOCK_DST)/board/service/service.go

proto_generate: 
	protoc \
    --go_out=./protos \
    --go_opt=module=protos \
    --go-grpc_out=./protos \
    --go-grpc_opt=module=protos \
    protos/proto/auth/auth.proto \
	protos/proto/feed/feed.proto \
	protos/proto/chat/chat.proto

test: mocks
	go test $(TESTED_DIRS) -coverprofile=$(COVERAGE_FILE)
	
cover: mocks test
	cat $(COVERAGE_FILE) | grep -v 'mock_' | grep -v 'docs' | grep -v 'test_utils' > cover.out
	go tool cover -func=cover.out