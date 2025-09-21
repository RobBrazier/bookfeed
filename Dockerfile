FROM golang:1.25.1-alpine AS build

WORKDIR /app

RUN apk add --no-cache nodejs npm

COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
	npm install

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
	go generate ./... && \
	CGO_ENABLED=0 go build -o main cmd/bookfeed/main.go

FROM alpine:3.22.1 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
EXPOSE ${PORT}
CMD ["./main"]


