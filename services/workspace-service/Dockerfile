FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY services/workspace-service/go.mod services/workspace-service/go.sum ./services/workspace-service/

WORKDIR /app/services/workspace-service
RUN go mod download

COPY . /app

RUN go build -ldflags="-s -w" -o /app/workspace-api ./cmd/api/main.go
RUN go build -ldflags="-s -w" -o /app/workspace-worker ./cmd/worker/main.go


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