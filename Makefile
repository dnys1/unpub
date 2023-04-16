VERSION := 2.1

.PHONY: build-docker
build-docker: build-server build-launcher

.PHONY: build-server
build-server:
	docker buildx build --load --target server -t dnys1/unpub:latest -t dnys1/unpub:v${VERSION} .

.PHONY: build-launcher
build-launcher:
	docker buildx build --load --target launcher -t dnys1/unpub-launcher:latest -t dnys1/unpub-launcher:v${VERSION} .

.PHONY: build-frontend
build-frontend:
	( cd web && webdev build )
	cp web/build/index.html cmd/server/build/index.html
	cp web/build/main.dart.js cmd/server/build/main.dart.js

.PHONY: build
build: build-frontend
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/unpub_linux_amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 go build -o bin/unpub_linux_arm64 ./cmd/server
	GOOS=darwin GOARCH=amd64 go build -o bin/unpub_darwin_amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 go build -o bin/unpub_darwin_arm64 ./cmd/server
	tar -czvf bin/unpub_linux_amd64.tar.gz -C bin unpub_linux_amd64
	tar -czvf bin/unpub_linux_arm64.tar.gz -C bin unpub_linux_arm64
	tar -czvf bin/unpub_darwin_amd64.tar.gz -C bin unpub_darwin_amd64
	tar -czvf bin/unpub_darwin_arm64.tar.gz -C bin unpub_darwin_arm64

.PHONY: test
test:
	go test -tags e2e -timeout 1m -v ./
