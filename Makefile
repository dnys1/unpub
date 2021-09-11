VERSION := 1.0

build: build-server build-launcher
	
build-server:
	docker build --target server -t dnys1/unpub:latest -t dnys1/unpub:v${VERSION} -f docker/latest/Dockerfile .

build-launcher:
	docker build --target launcher -t dnys1/unpub-launcher:latest -t dnys1/unpub-launcher:v${VERSION} -f docker/latest/Dockerfile .
