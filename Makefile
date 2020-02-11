test:
	go test ./... -v -cover -coverprofile=cover.out
	go tool cover -html=cover.out