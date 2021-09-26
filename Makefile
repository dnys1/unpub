VERSION := 2.0

.PHONY: build-docker
build-docker: build-server build-launcher

.PHONY: build-server
build-server:
	docker buildx build --target server -t dnys1/unpub:latest -t dnys1/unpub:v${VERSION} .

.PHONY: build-launcher
build-launcher:
	docker buildx build --target launcher -t dnys1/unpub-launcher:latest -t dnys1/unpub-launcher:v${VERSION} .

.PHONY: build-frontend
build-frontend:
	( cd web && webdev build )
	cp -R web/build .

.PHONY: build
build: build-frontend
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o bin/launcher_linux_amd64 .
	go build -o bin/launcher_darwin_amd64 .
	tar -czvf bin/launcher_linux_amd64.tar.gz -C bin launcher_linux_amd64
	tar -czvf bin/launcher_darwin_amd64.tar.gz -C bin launcher_darwin_amd64
