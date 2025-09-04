# Hardcover RSS

A Go application that generates RSS feeds for Hardcover.app, a book tracking and review platform. This service provides RSS feeds for recent book releases, author releases, series releases, and personalized user feeds in multiple formats (RSS, Atom, and JSON).

## Features

- Recent book releases feed
- Author-specific releases feed
- Series-specific releases feed
- Personalized user feeds based on reading history
- Multiple output formats: RSS, Atom, and JSON
- Rate-limited API endpoints for public access

## Prerequisites

- Go 1.25+
- Hardcover API token (for development)

## Getting Started

These instructions will get you a copy of the project up and running on your local machine.

### Environment Setup

Copy the example environment file and configure your settings:

```bash
cp .env.example .env
```

Edit the `.env` file to set your configuration:
- `PORT`: The port to run the server on (default: 8000)
- `HARDCOVER_TOKEN`: Your Hardcover API token (required for development)

### Installation

Install dependencies:

```bash
go mod tidy
```

### Running the Application

Run directly with Go:

```bash
go run cmd/api/main.go
```

Or use Task (if installed):

```bash
task run
```

For development with live reload:

```bash
task dev
```

### Building

To build a binary:

```bash
go build -o dist/hardcover-feed cmd/api/main.go
```

Or with Task:

```bash
task build
```

### Testing

Run tests:

```bash
go test ./... -v
```

Or with Task:

```bash
task test
```

## Usage

Once running, the application exposes the following endpoints:

### Recent Releases
- `GET /recent` - Recent releases in RSS format
- `GET /recent.rss` - Recent releases in RSS format
- `GET /recent.atom` - Recent releases in Atom format
- `GET /recent.json` - Recent releases in JSON format

### Author Releases
- `GET /author/{author}` - Specific author's releases in RSS format
- `GET /author/{author}.rss` - Specific author's releases in RSS format
- `GET /author/{author}.atom` - Specific author's releases in Atom format
- `GET /author/{author}.json` - Specific author's releases in JSON format

### Series Releases
- `GET /series/{series}` - Specific series' releases in RSS format
- `GET /series/{series}.rss` - Specific series' releases in RSS format
- `GET /series/{series}.atom` - Specific series' releases in Atom format
- `GET /series/{series}.json` - Specific series' releases in JSON format

### Personalized User Feeds
- `GET /me/{username}` - Personalized releases based on user's reading history in RSS format
- `GET /me/{username}.rss` - Personalized releases based on user's reading history in RSS format
- `GET /me/{username}.atom` - Personalized releases based on user's reading history in Atom format
- `GET /me/{username}.json` - Personalized releases based on user's reading history in JSON format
- `GET /me/{username}?filter=author` - Filter to only show author releases
- `GET /me/{username}?filter=series` - Filter to only show series releases

### Development Tasks

Update the GraphQL schema from the Hardcover API:

```bash
task download-schema
```

Generate Go code from GraphQL schema:

```bash
task generate
```

## Deployment

The application can be deployed as a standalone binary or Docker container. It requires the `PORT` environment variable to be set.

### Docker

Build the Docker image:

```bash
docker build -t hardcover-feed .
```

Run with Docker:

```bash
docker run -p 8000:8000 hardcover-feed
```
