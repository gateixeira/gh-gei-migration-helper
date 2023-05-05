clean:
	rm -rf dist/
build:
	go build -o dist/gei-migration-helper main.go
snapshot:
	goreleaser release --snapshot
release:
	goreleaser release