FROM golang:1.25.2-alpine@sha256:06cdd34bd531b810650e47762c01e025eb9b1c7eadd191553b91c9f2d549fae8 AS build

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


