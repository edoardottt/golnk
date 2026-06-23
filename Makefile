REPO=github.com/edoardottt/golnk

remod:
	@rm -rf go.*
	@go mod init ${REPO}
	@go get ./...
	@go mod tidy -v
	@echo "Done."

update:
	@go get -u ./...
	@go mod tidy -v
	@echo "Done."

lint:
	@golangci-lint run

linux:
	@go build -o golnk ./cmd/golnk
	@sudo mv golnk /usr/local/bin/
	@echo "Done."

unlinux:
	@sudo rm -rf /usr/local/bin/golnk
	@echo "Done."

test:
	@go test -race ./...