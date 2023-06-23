# syntax=docker/dockerfile:1
FROM golang:1.20.5-bullseye AS build

WORKDIR /work/backend

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/cache/download \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build .

FROM scratch AS release

COPY --from=build /work/backend/backend /usr/bin/backend

EXPOSE 8080

ENTRYPOINT [ "/usr/bin/backend" ]
CMD [ "--host=0.0.0.0" ]
