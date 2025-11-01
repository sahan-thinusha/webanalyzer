FROM golang:1.25.3-alpine AS builder

RUN apk add --no-cache git bash ca-certificates
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o webanalyzer ./cmd/webanalyzer

FROM alpine:3.22

RUN apk add --no-cache ca-certificates && update-ca-certificates
RUN adduser -D -g '' appuser
USER appuser

WORKDIR /app

COPY --from=builder /app/webanalyzer .
COPY --from=builder /app/.env .env

EXPOSE 8080 8081 6061

ENTRYPOINT ["./webanalyzer"]
