FROM golang:1.23.0 AS builder
LABEL authors="sakojpa"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -o /tasker main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /tasker /app/tasker
COPY web /app/web

ENTRYPOINT ["/app/tasker"]