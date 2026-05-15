FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY service/go.mod service/go.sum ./
RUN go mod download
COPY service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /url-shortener .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /url-shortener /url-shortener
EXPOSE 8080
ENTRYPOINT ["/url-shortener"]
