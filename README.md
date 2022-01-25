# Unpub

Self-contained Unpub server which can be run in-memory or from the local filesystem. Includes a launcher which will seed the server with packages from a git repository.

## Usage

```bash
$ unpub
```

The server is controlled by the following flags:

| Flag              | Function                            | Default                   |
| ----------------- | ----------------------------------- | ------------------------- |
| `-port`           | The local port to run unpub on      | 4000                      |
| `-memory`         | Whether to run the server in-memory | `false`                   |
| `-path`           | Where to store files                | Temp dir                  |
| `-uploader-email` | The default uploader email to use   | test@example.com          |
| `-launch`         | Whether to run the launcher         | `false`                   |
| `-addr`           | The address Unpub is running on     | `http://localhost:{PORT}` |

## Build

> Requires Go 1.16 or higher, and Dart 2.14 or higher

To build the server with bundled launcher, run the following commands:

```sh
#!/bin/sh

# Build the frontend
pushd web
dart pub get
dart run build_runner build --delete-conflicting-outputs
dart pub global activate webdev
webdev build
popd

cp web/build/index.html cmd/server/build/index.html
cp web/build/main.dart.js cmd/server/build/main.dart.js

# Build the server
go build -o unpub ./cmd/server
```

To build the standalone launcher:

```sh
$ go build -o launch ./cmd/launcher
```

## Launcher

The tool looks for Dart packages by recursively walking the file tree and searching for `pubspec.yaml` files which do not specify `publish_to: none`.

The launcher is controlled with the following environment variables:

| Variable        | Function                     | Default        |
| --------------- | ---------------------------- | -------------- |
| `UNPUB_PORT`    | The local port running unpub | N/A (required) |
| `UNPUB_GIT_URL` | The git url to clone         | N/A (required) |
| `UNPUB_GIT_REF` | The git ref to clone         | main           |
