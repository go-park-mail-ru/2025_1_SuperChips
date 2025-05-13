GO := go
MOCKGEN=mockgen
NAME := main.exe
COVERAGE_FILE=coverage.out
MOCK_DST=./mocks
DOMAIN_FLDR := domain
REST_FLDR := internal/rest
TESTED_DIRS := ./$(REST_FLDR)/... ./$(DOMAIN_FLDR)/... ./internal/repository/...

.PHONY : test mocks

mocks:
	@mkdir -p $(MOCK_DST)/pin $(MOCK_DST)/user
	$(MOCKGEN) -source=./pin/service.go -destination=$(MOCK_DST)/pin/repository/repository.go
	$(MOCKGEN) -source=./auth/service.go -destination=$(MOCK_DST)/user/repository/repository.go
	$(MOCKGEN) -source=./profile/service.go -destination=$(MOCK_DST)/profile/repository/repository.go
	$(MOCKGEN) -source=./$(REST_FLDR)/profile.go -destination=$(MOCK_DST)/profile/service/service.go
	$(MOCKGEN) -source=./board/service.go -destination=$(MOCK_DST)/board/repository/repository.go
	$(MOCKGEN) -source=./internal/grpc/auth.go -destination=$(MOCK_DST)/user/service/service.go
	$(MOCKGEN) -source=./like/service.go -destination=$(MOCK_DST)/like/repository/repository.go
	$(MOCKGEN) -source=./$(REST_FLDR)/like.go -destination=$(MOCK_DST)/like/service/service.go
	$(MOCKGEN) -source=./$(REST_FLDR)/board.go -destination=$(MOCK_DST)/board/service/service.go
	$(MOCKGEN) -source=./protos/gen/auth/auth_grpc.pb.go -destination=$(MOCK_DST)/auth/grpc/client.go
	$(MOCKGEN) -source=./protos/gen/feed/feed_grpc.pb.go -destination=$(MOCK_DST)/feed/grpc/client.go
	$(MOCKGEN) -source=./protos/gen/chat/chat_grpc.pb.go -destination=$(MOCK_DST)/chat/grpc/client.go
	$(MOCKGEN) -source=./$(REST_FLDR)/search.go -destination=$(MOCK_DST)/search/service/service.go
	$(MOCKGEN) -source=./$(REST_FLDR)/subscription.go -destination=$(MOCK_DST)/subscription/service/service.go
	$(MOCKGEN) -source=./internal/grpc/feed.go -destination=$(MOCK_DST)/feed/service/service.go


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
	cat $(COVERAGE_FILE) | grep -v 'mock_' | grep -v 'docs' | grep -v 'test_utils' | grep -v 'gen' > cover.out
	go tool cover -func=cover.out

easyjson:
	easyjson $(DOMAIN_FLDR)/auth.go \
	$(DOMAIN_FLDR)/board.go \
	$(DOMAIN_FLDR)/chat.go \
	$(DOMAIN_FLDR)/feed.go \
	$(DOMAIN_FLDR)/like.go \
	$(DOMAIN_FLDR)/user.go \
	$(DOMAIN_FLDR)/pincrud.go \
	$(REST_FLDR)/helper.go \
	$(REST_FLDR)/board.go \
	$(REST_FLDR)/chat.go \
	$(REST_FLDR)/profile.go \
	$(REST_FLDR)/subscription.go

easyjson_stub:
	easyjson -stub $(DOMAIN_FLDR)/auth.go \
	$(DOMAIN_FLDR)/board.go \
	$(DOMAIN_FLDR)/chat.go \
	$(DOMAIN_FLDR)/feed.go \
	$(DOMAIN_FLDR)/like.go \
	$(DOMAIN_FLDR)/user.go \
	$(DOMAIN_FLDR)/pincrud.go \
	$(REST_FLDR)/helper.go \
	$(REST_FLDR)/board.go \
	$(REST_FLDR)/chat.go \
	$(REST_FLDR)/profile.go \
	$(REST_FLDR)/subscription.go
