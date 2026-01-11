# paperless-cli

Use paperless-cli when the user wants to manage documents in their Paperless-ngx instance. This includes listing, searching, uploading, downloading documents, and managing tags, correspondents, and document types.

## Auth

Set environment variables (recommended):
```bash
export PAPERLESS_URL="https://paperless.example.com"
export PAPERLESS_TOKEN="your-api-token"
```

Or save to config:
```bash
paperless config set-url https://paperless.example.com
paperless config set-token your-api-token
```

## Documents

```bash
paperless documents list                    # List recent documents
paperless documents list --limit 10         # Limit results
paperless documents list --query "invoice"  # Filter by search term
paperless documents list --tag bills        # Filter by tag
paperless documents search "contract 2024"  # Full-text search
paperless documents get <id>                # Get document details
paperless documents content <id>            # Get extracted text
paperless documents upload file.pdf         # Upload document
paperless documents download <id>           # Download document
paperless documents edit <id> --title "New" # Edit metadata
paperless documents delete <id>             # Delete document
```

## Tags, Correspondents, Types

```bash
paperless tags list                         # List all tags
paperless tags create "receipts"            # Create tag
paperless correspondents list               # List correspondents
paperless correspondents create "ACME"      # Create correspondent
paperless types list                        # List document types
paperless types create "Invoice"            # Create document type
```

## PDF Utilities

```bash
paperless pdf read document.pdf             # Extract text from local PDF
paperless pdf info document.pdf             # Show PDF metadata
```

## Tasks

```bash
paperless tasks status <task-id>            # Check upload task status
```

## Options

| Flag | Description |
|------|-------------|
| `-h, --help` | Show help |
| `--version` | Print version |
| `-q, --quiet` | Suppress non-essential output |
| `--json` | Output as JSON (for scripting) |
| `--no-color` | Disable color output |
| `-u, --url` | Override server URL |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PAPERLESS_URL` | Paperless server URL |
| `PAPERLESS_TOKEN` | API authentication token |

## Examples

### List recent documents

```bash
$ paperless documents list --limit 5
ID  TITLE                                     CREATED     TAGS
95  Kirchenaustritt                           2025-11-03  0 tags
96  Mietvertrag Judith                        2025-11-03  0 tags
94  93015_27409_Uebertragungsprotokoll_EU...  2025-03-16  0 tags
64  Steuernummer 2025                         2025-03-12  0 tags
61  Steuerbescheid für 2023                   2025-03-04  0 tags

Showing 5 of 95 documents
```

### Search documents

```bash
$ paperless documents search "invoice"
ID  TITLE               CREATED
23  Invoice March 2024  2024-03-15
18  Invoice Feb 2024    2024-02-10

Found 2 documents
```

### Get document details

```bash
$ paperless documents get 95
ID:           95
Title:        Kirchenaustritt
Created:      2025-11-03
Added:        2025-11-03 09:10:52
Modified:     2025-11-03 09:10:52
Original:     Kirchenaustritt.pdf
```

### Upload a document

```bash
$ paperless documents upload invoice.pdf --title "January Invoice" --tag bills
Uploading invoice.pdf...
Uploaded invoice.pdf (task: abc-123-def)
```

### Check upload task status

```bash
$ paperless tasks status abc-123-def
Task ID:     abc-123-def
Status:      SUCCESS
Type:        file
File:        invoice.pdf
Created:     2024-01-11T18:37:12.449939Z
Completed:   2024-01-11T18:37:15.442070Z
Result:      Success. New document id 97 created
Document:    97
```

### Download a document

```bash
$ paperless documents download 97 -o ~/Downloads/invoice.pdf
Downloaded to /Users/me/Downloads/invoice.pdf (45678 bytes)
```

### Get document text content

```bash
$ paperless documents content 95
[Extracted text from the document...]
```

### Read local PDF

```bash
$ paperless pdf read document.pdf
This is the extracted text content from the PDF file...
```

### JSON output for scripting

```bash
$ paperless documents list --json --limit 2
{
  "count": 95,
  "results": [
    {"id": 95, "title": "Kirchenaustritt", "created_date": "2025-11-03", ...},
    {"id": 96, "title": "Mietvertrag Judith", "created_date": "2025-11-03", ...}
  ]
}
```

### List tags

```bash
$ paperless tags list
ID  NAME      COLOR    DOCS
1   bills     #ff0000  12
2   receipts  #00ff00  8
3   contracts #0000ff  5
```

### Create a tag

```bash
$ paperless tags create "important" --color "#ff0000"
Created tag 4: important
```

## Notes

- API token can be obtained from Paperless-ngx admin panel
- Uploaded documents are processed asynchronously; check task status for completion
- Use `--json` flag for machine-readable output
- Config stored in `~/.config/paperless-cli/config.yaml`
- Tags, correspondents, and types can be specified by name or ID in upload/edit commands
