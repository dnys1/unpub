build: build-server build-launcher
	
build-server:
	docker build --target server -t dnys1/unpub:latest -f docker/latest/Dockerfile .

build-launcher:
	docker build --target launcher -t dnys1/unpub-launcher:latest -f docker/latest/Dockerfile .