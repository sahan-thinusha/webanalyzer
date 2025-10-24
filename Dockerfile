FROM golang:1.25.3-alpine AS builder

RUN apk add --no-cache git bash
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o webanalyzer ./cmd/webanalyzer

FROM alpine:3.22

RUN adduser -D -g '' appuser
USER appuser

WORKDIR /app

COPY --from=builder /app/webanalyzer .

ENV BASIC_AUTH_USER=admin
ENV BASIC_AUTH_PASS=pw12345
ENV PORT=8080

EXPOSE 8080

CMD ["./webanalyzer"]