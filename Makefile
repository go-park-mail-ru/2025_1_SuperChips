GO := go
NAME := main.exe

build : cmd/main.go
	$(GO) build -o $(NAME) $<

.PHONY : build