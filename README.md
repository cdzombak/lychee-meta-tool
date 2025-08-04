# Lychee Meta Tool

A web-based tool for quickly adding titles to photos in Lychee photo libraries. Automatically identifies photos with generic camera names (IMG_1234, CD5_5678, etc.) and provides an efficient interface for adding meaningful titles.

## Features

- **Smart Detection**: Automatically finds photos with generic camera names, UUIDs, or empty titles
- **Album Filtering**: Work on photos from specific albums only
- **Keyboard Navigation**: Cmd+J/K for previous/next photo
- **Single Binary Deployment**: All frontend assets embedded
- **Multi-Database Support**: MySQL, PostgreSQL, SQLite

## Quick Start

1. **Build**:
   ```bash
   git clone https://github.com/cdzombak/lychee-meta-tool.git
   cd lychee-meta-tool
   make build
   ```

2. **Configure**:
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your database settings
   ```

3. **Run**:
   ```bash
   ./lychee-meta-tool -config config.yaml
   ```

4. **Open**: http://localhost:8080

## Configuration

```yaml
database:
  type: mysql  # mysql, postgres, or sqlite
  host: localhost
  port: 3306
  user: lychee
  password: your_password
  database: lychee

lychee_base_url: "https://your-lychee-instance.com"

server:
  port: 8080
```

## Usage

1. Select photos from the filmstrip at the top
2. Use the album filter in the right panel to focus on specific albums
3. Add titles in the right panel editor
4. Press Enter to save and move to next photo
5. Use Cmd+J/K for keyboard navigation

## Development

```bash
# Frontend development with hot reload
cd frontend && npm run dev

# Backend development
go build -o lychee-meta-tool .
./lychee-meta-tool -config config.yaml
```

## Photo Detection

Identifies photos needing titles by matching patterns:
- 3-character camera prefixes: `CD5_1234`, `IMG_5678`, `DSZ_9012`
- UUID-based filenames
- Screenshot and timestamp patterns
- Empty or null titles

## License

MIT License