VERSION := 1.0

.PHONY: build-docker
build-docker: build-server build-launcher

.PHONY: build-server
build-server:
	docker build --target server -t dnys1/unpub:latest -t dnys1/unpub:v${VERSION} -f docker/latest/Dockerfile .

.PHONY: build-launcher
build-launcher:
	docker build --target launcher -t dnys1/unpub-launcher:latest -t dnys1/unpub-launcher:v${VERSION} -f docker/latest/Dockerfile .

.PHONY: build
build:
	mkdir -p build && \
	GOOS=linux GOARCH=amd64 go build -o build/launcher_linux_amd64 . && \
	go build -o build/launcher_darwin_amd64 . && \
	tar -czvf build/launcher_linux_amd64.tar.gz -C build launcher_linux_amd64 && \
	tar -czvf build/launcher_darwin_amd64.tar.gz -C build launcher_darwin_amd64
