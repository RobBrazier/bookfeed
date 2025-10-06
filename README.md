# BookFeed

A Go application that generates standardized RSS/Atom/JSON feeds for book tracking and review platforms.

## Supported Providers

### Hardcover.app (Current)
This service provides feeds for recent book releases, author releases, series releases, and personalized user feeds from Hardcover.app.

### Future Provider Support
The application is architected to potentially support additional book tracking platforms in the future.

## Features

- Multiple output formats: RSS, Atom, and JSON
- Rate-limited API endpoints for public access

### Hardcover
- Recent book releases feed
- Author-specific releases feed
- Series-specific releases feed
- Personalized user feeds based on reading history

## Prerequisites

- Go 1.25+
- Provider API token (for development - Hardcover token currently required)

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

Or use Just (if installed):

```bash
just run
```

For development with live reload:

```bash
just dev
```

### Building

To build a binary:

```bash
mkdir dist
go build -o dist/bookfeed cmd/api/main.go
```

Or with Just:

```bash
just build
```

### Testing

Run tests:

```bash
go test ./... -v
```

Or with Just:

```bash
just test
```

## Usage

Once running, the application exposes the following endpoints:

### Recent Releases
- `GET /hc/recent.atom` - Recent releases in Atom format
- `GET /hc/recent.rss` - Recent releases in RSS format
- `GET /hc/recent.json` - Recent releases in JSON format

### Author Releases
- `GET /hc/author/{author}.atom` - Specific author's releases in Atom format
- `GET /hc/author/{author}.rss` - Specific author's releases in RSS format
- `GET /hc/author/{author}.json` - Specific author's releases in JSON format

### Series Releases
- `GET /hc/series/{series}.atom` - Specific series' releases in Atom format
- `GET /hc/series/{series}.rss` - Specific series' releases in RSS format
- `GET /hc/series/{series}.json` - Specific series' releases in JSON format

### Personalized User Feeds
- `GET /hc/me/{username}.atom` - Personalized releases based on user's reading history in Atom format
- `GET /hc/me/{username}.rss` - Personalized releases based on user's reading history in RSS format
- `GET /hc/me/{username}.json` - Personalized releases based on user's reading history in JSON format
- `GET /hc/me/{username}.atom?filter=author` - Filter to only show author releases
- `GET /hc/me/{username}.atom?filter=series` - Filter to only show series releases

### Development Tasks

Update the GraphQL schema from the Hardcover API:

```bash
just downloadSchema
```

Generate Go code from GraphQL schema:

```bash
just generate
```

## Deployment

The application can be deployed as a standalone binary or Docker container. It requires the `PORT` environment variable to be set.

### Docker

Build the Docker image:

```bash
docker build -t bookfeed .
```

Run with Docker:

```bash
docker run -p 8000:8000 bookfeed
```
