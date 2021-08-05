# Unpub Launcher

Launch an Unpub server locally with Docker Compose and seed it with packages in a git repository. The tool looks for Dart packages by recursively walking the file tree and searching for `pubspec.yaml` files which do not specify `publish_to: none`.

## Usage

The launcher is controlled with the following environment variables:

| Variable | Function | Default |
| -------- | -------- | ------- |
| `UNPUB_PORT` | The port to run unpub on | N/A (required) |
| `UNPUB_GIT_URL` | The git url to clone | N/A (required) |
| `UNPUB_GIT_REF` | The git ref to clone | main |

```bash
$ curl -L https://raw.githubusercontent.com/dnys1/unpub-launcher/main/docker/latest/docker-compose.yml -o docker-compose.yml
$ UNPUB_PORT=8000 \
  UNPUB_GIT_URL=https://github.com/dnys1/my-dart-packages \
  docker compose up
```