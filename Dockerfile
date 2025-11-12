FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /api ./

EXPOSE 8080

CMD ["./api"]
