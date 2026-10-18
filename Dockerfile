# Build the Go application
FROM golang:1.27.1-alpine AS builder-go

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY backend/ ./backend/

RUN go build ./backend/main.go

# Build the Node.js application
FROM node:26.10-alpine AS builder-node

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./

RUN npm ci

COPY frontend/ ./

RUN npm run build

# Final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /app

COPY --from=builder-go /app/main .

COPY --from=builder-node /app/dist ./frontend/dist

EXPOSE 3030

CMD ["./main"]
