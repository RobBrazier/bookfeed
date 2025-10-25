FROM golang:1.25.3-alpine@sha256:aee43c3ccbf24fdffb7295693b6e33b21e01baec1b2a55acc351fde345e9ec34 AS build

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

FROM alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
EXPOSE ${PORT}
CMD ["./main"]


