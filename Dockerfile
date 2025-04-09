FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/fileuploader cmd/api/main.go

FROM alpine:latest

RUN mkdir -p /app/uploads/temp /app/uploads/files

WORKDIR /app

COPY --from=builder /app/fileuploader .
COPY .env .

EXPOSE 8080

CMD ["./fileuploader"]
