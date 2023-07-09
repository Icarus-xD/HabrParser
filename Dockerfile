# Stage 1
FROM golang:1.20 AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -o HabrParser cmd/main.go

# Stage 2
# FROM alpine:3.18
FROM scratch

WORKDIR /app

COPY --from=builder /app/HabrParser /app/HabrParser

COPY .env .

CMD ["./HabrParser"]