# Build the unpub server
FROM --platform=amd64 dart:stable AS build-server

RUN dart pub global activate unpub && \
    cd /root/.pub-cache/hosted/pub.dartlang.org && \
    cp -r $(ls | grep unpub) /unpub
    
WORKDIR /unpub
RUN mkdir -p build && \
    dart pub get && \
    dart compile exe -o build/server bin/unpub.dart

# Build the launcher tool
FROM --platform=$BUILDPLATFORM golang:1.16 AS build-launcher

WORKDIR /launcher

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
    GOARCH=$(echo $TARGETPLATFORM | cut -d / -f 2) go build -o bin/launch .

# Ouput the server image
FROM ubuntu:hirsute AS server

RUN apt update && apt install -y ca-certificates

WORKDIR /unpub
COPY --from=build-server /unpub/build/server .

ENTRYPOINT [ "/unpub/server" ]

HEALTHCHECK --interval=5s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${UNPUB_PORT} || exit 1

# Output the launcher image
FROM ubuntu:hirsute AS launcher

RUN apt update && apt install -y ca-certificates

COPY --from=build-launcher /launcher/bin/launch /usr/local/bin

ENTRYPOINT [ "launch" ]
