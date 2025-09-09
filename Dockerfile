FROM golang:1.25.1-alpine AS build

WORKDIR /app

RUN apk add --no-cache nodejs npm && \
	npm install -g bun

COPY package.json bun.lock ./
RUN bun install

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go generate ./... && \
	CGO_ENABLED=0 go build -o main cmd/api/main.go

FROM alpine:3.22.1 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
EXPOSE ${PORT}
CMD ["./main"]


