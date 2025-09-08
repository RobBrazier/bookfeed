FROM oven/bun:alpine AS node

COPY package.json .
COPY bun.lock .

RUN bun install

FROM golang:1.25.1-alpine AS build

WORKDIR /app

RUN apk add --no-cache nodejs npm

COPY --from=node /home/bun/app/node_modules /app/node_modules

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go generate ./... && \
	go build -buildvcs -o main cmd/api/main.go

FROM alpine:3.22.1 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
EXPOSE ${PORT}
CMD ["./main"]


