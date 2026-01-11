# paperless-cli

A command-line interface for [Paperless-ngx](https://docs.paperless-ngx.com/), the open-source document management system.

## Installation

```bash
# From source
go install github.com/julianfbeck/paperless-cli@latest

# Or clone and build
git clone https://github.com/julianfbeck/paperless-cli.git
cd paperless-cli
go build -o paperless .
```

## Configuration

Set your Paperless-ngx server URL and API token:

```bash
# Using environment variables (recommended)
export PAPERLESS_URL="https://paperless.example.com"
export PAPERLESS_TOKEN="your-api-token"

# Or save to config file
paperless config set-url https://paperless.example.com
paperless config set-token your-api-token
```

Get your API token from the Paperless-ngx admin panel at `/admin/authtoken/tokenproxy/`.

## Usage

### Documents

```bash
# List documents
paperless documents list
paperless documents list --limit 10 --query "invoice"

# Search
paperless documents search "contract 2024"

# Get details
paperless documents get 123

# Upload
paperless documents upload invoice.pdf --title "January Invoice"

# Download
paperless documents download 123 -o ~/Downloads/doc.pdf

# Get extracted text
paperless documents content 123

# Edit
paperless documents edit 123 --title "New Title" --add-tag important

# Delete
paperless documents delete 123
```

### Tags, Correspondents, Document Types

```bash
# List
paperless tags list
paperless correspondents list
paperless types list

# Create
paperless tags create "receipts" --color "#ff0000"
paperless correspondents create "ACME Corp"
paperless types create "Invoice"

# Edit
paperless tags edit 1 --name "new-name"

# Delete
paperless tags delete 1 --force
```

### PDF Utilities

```bash
# Extract text from local PDF
paperless pdf read document.pdf

# Get PDF info
paperless pdf info document.pdf
```

### Tasks

```bash
# Check upload task status
paperless tasks status abc-123-def
```

## Options

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |
| `-q, --quiet` | Suppress non-essential output |
| `--no-color` | Disable color output |
| `-u, --url` | Override server URL |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PAPERLESS_URL` | Server URL |
| `PAPERLESS_TOKEN` | API token |

## Development

```bash
# Build
go build .

# Run tests (requires PAPERLESS_URL and PAPERLESS_TOKEN)
go test -tags=local ./...

# Generate test PDF
go run testdata/generate_test_pdf.go
```

## License

MIT
