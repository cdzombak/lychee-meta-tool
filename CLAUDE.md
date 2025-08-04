# Lychee Meta Tool - Development Notes

## Project Overview
Quickly find & edit untitled photos in your Lychee photo library.

## Architecture
- **Backend**: Go with embedded frontend assets using `go:embed`
- **Frontend**: Vue.js with Vite
- **Database**: MySQL, PostgreSQL, or SQLite support
- **Config**: YAML configuration file

## Photo Title Detection
Photos needing titles are identified by:
- Empty strings or null values for title
- 3-character camera prefixes: `^[A-Za-z0-9]{3}_[0-9]+(\\.\\w+)?$` (e.g., IMG_1234, CD5_5678, CDZ_9012)
- Other camera patterns: P-series, timestamps, WhatsApp images, screenshots
- UUID patterns (with or without file extensions)

## Key Features
- **Album Filtering**: Filter photos by specific albums (only shows albums with photos needing titles)
- **Keyboard Navigation**: Cmd+J/K for previous/next photo
- **Auto-focus**: Title field auto-focuses and selects text on photo selection
- **Toast Notifications**: User feedback displayed in bottom-right corner
- **Real-time Updates**: Photos disappear from list after titles are saved

## API Endpoints
- `GET /api/photos/needsmetadata` - Photos without proper titles (supports `?album_id=` filter)
- `GET /api/photos/:id` - Single photo details
- `PUT /api/photos/:id` - Update photo metadata
- `GET /api/albums` - All normal albums
- `GET /api/albums/withphotocounts` - Albums that contain photos needing titles

## Configuration
```yaml
database:
  type: mysql|postgres|sqlite
  host: localhost
  port: 3306
  user: lychee
  password: password
  database: lychee
lychee_base_url: "https://your-lychee-instance.com"
server:
  port: 8080
  cors:
    allowed_origins:
      - http://localhost:5173
```

## Development Commands
```bash
# Build everything
make build

# Backend only
go build -o lychee-meta-tool .

# Frontend only
cd frontend && npm run build

# Development
cd frontend && npm run dev
```