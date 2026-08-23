# syntax=docker/dockerfile:1.7
FROM golang:1.24-alpine AS compile
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -o /out/resume-api ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -o /out/resume-migrate ./cmd/migrate

FROM alpine:3.22 AS runtime
RUN addgroup -S resume && adduser -S -G resume -h /var/lib/resume-workspace resume
WORKDIR /opt/resume
COPY --from=compile /out/resume-api /usr/local/bin/resume-api
COPY --from=compile /out/resume-migrate /usr/local/bin/resume-migrate
COPY migrations ./migrations
RUN mkdir -p /var/lib/resume-workspace/files && chown -R resume:resume /var/lib/resume-workspace
USER resume
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/resume-api"]
