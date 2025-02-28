GO := go

main : cmd/main.go
	$(GO) build $<
