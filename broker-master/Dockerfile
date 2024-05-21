FROM docker.arvancloud.ir/library/golang:1.21-alpine
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY . .
RUN go install -v ./...
EXPOSE 8080
CMD go run main.go