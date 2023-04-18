build:
	go build -o bin/gei-migration-helper main.go
compile:
	GOOS=linux GOARCH=arm go build -o bin/gei-migration-helper-linux-arm main.go
	GOOS=linux GOARCH=amd64 go build -o bin/gei-migration-helper-linux-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o bin/gei-migration-helper-windows-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/gei-migration-helper-darwin-arm64 main.go