set dotenv-load := true
set unstable := true

name := "bookfeed"
main := "cmd" / name / "main.go"
out := "dist" / name
root := justfile_directory()
run := if env("SECRETS_COMMAND", "") != "" { env("SECRETS_COMMAND") + " -- " } else { "" }
air := if which("air") != "" { which("air") } else { "go run github.com/air-verse/air@latest" }
hardcoverApi := "https://api.hardcover.app/v1/graphql"
export CGO_ENABLED := "0"

# List available recipes
@list:
    just --list

# Install frontend dependencies and setup pre-commit hooks
install:
    npm install
    -npx lefthook install

# Run formatting
fmt *args:
    golangci-lint fmt {{ args }}

# Run linting (with fixes enabled)
lint *args:
    golangci-lint run --fix {{ args }}

# Clean temporary files/folders
clean:
    -rm -r {{ root }}/dist

# Run code generation
generate:
    go generate ./...

# Build the app
build OUT=out: generate
    go build -o {{ OUT }} {{ main }}

# Run the app
run:
    {{ run }}go run {{ main }}

# Run the app in dev mode (auto-restart) with air
dev:
    {{ run }}{{ air }}

# Run tests
test:
    go test ./... -v

# Run TemplUI commands
templui *args:
    go run github.com/templui/templui/cmd/templui@latest {{ args }}

# Update all TemplUI components
templuiUpdate:
    just templui -f add $(ls internal/view/components)

# Download the Hardcover GraphQL Schema
schema:
    go run github.com/benweint/gquil/cmd/gquil@latest introspection generate-sdl {{ hardcoverApi }} -H "Authorization: {{ env('HARDCOVER_TOKEN') }}" > internal/schema/schema.gql
