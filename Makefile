VERSION := 2.0

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
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o bin/launcher_linux_amd64 .
	go build -o bin/launcher_darwin_amd64 .
	tar -czvf bin/launcher_linux_amd64.tar.gz -C bin launcher_linux_amd64
	tar -czvf bin/launcher_darwin_amd64.tar.gz -C bin launcher_darwin_amd64
