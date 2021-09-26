# Build the frontend
FROM --platform=amd64 dart:2.14 AS build-frontend

WORKDIR /unpub

RUN dart pub global activate webdev

COPY . .

WORKDIR /unpub/web 
RUN dart pub get && \
    dart pub global run webdev build

# Build the server
FROM --platform=$BUILDPLATFORM golang:1.17 AS build-server

WORKDIR /unpub

# Opt-out of proxy
RUN go env -w GOPROXY=direct

# Improve build caching by downloading dependencies
# before copying source files
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
COPY --from=build-frontend /unpub/web/build ./cmd/server/build

ARG TARGETPLATFORM
RUN mkdir -p build && \
    GOARCH=$(echo $TARGETPLATFORM | cut -d / -f 2) go build -o bin/server ./cmd/server

# Build the launcher tool
FROM --platform=$BUILDPLATFORM golang:1.17 AS build-launcher

WORKDIR /unpub

# Opt-out of proxy
RUN go env -w GOPROXY=direct

# Improve build caching by downloading dependencies
# before copying source files
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ARG TARGETPLATFORM
RUN mkdir -p build && \
    GOARCH=$(echo $TARGETPLATFORM | cut -d / -f 2) go build -o bin/launcher ./cmd/launcher

# Ouput the server image
FROM ubuntu:hirsute AS server

RUN apt update && apt install -y ca-certificates

WORKDIR /unpub
COPY --from=build-server /unpub/bin/server /usr/local/bin/unpub

ENTRYPOINT [ "unpub" ]

HEALTHCHECK --interval=5s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${UNPUB_PORT} || exit 1

# Output the launcher image
FROM ubuntu:hirsute AS launcher

RUN apt update && apt install -y ca-certificates

COPY --from=build-launcher /unpub/bin/launcher /usr/local/bin/launch

ENTRYPOINT [ "launch" ]
