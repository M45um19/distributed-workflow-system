
FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go version && go mod download

COPY . .

RUN go build -ldflags="-s -w" -o workspace-api ./cmd/api/main.go
RUN go build -ldflags="-s -w" -o workspace-worker ./cmd/worker/main.go



FROM alpine:3.19

RUN adduser -D -g '' appuser
WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/workspace-api .
COPY --from=builder /app/workspace-worker .

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

CMD ["./workspace-api"]