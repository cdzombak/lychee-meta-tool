# Lychee Meta Tool - Architecture Reference

## Overview
Technical reference for the Lychee Meta Tool, a web application for managing photo titles in Lychee photo libraries.

## Architecture

### Backend (Go)
- RESTful API server using Go standard library
- Direct database connections (MySQL/PostgreSQL/SQLite)
- YAML configuration
- Frontend assets embedded using `go:embed` for single-binary deployment

### Frontend (Vue.js)
- Vue 3 with Composition API
- Pinia for state management
- Vite for build tooling
- Build artifacts embedded in Go binary

## Project Structure
```
lychee-meta-tool/
├── main.go                  # Entry point with embedded frontend
├── backend/
│   ├── config/config.go     # Configuration loading
│   ├── db/
│   │   ├── connection.go    # Database connection management
│   │   └── queries.go       # SQL queries with pattern matching
│   ├── handlers/
│   │   ├── photos.go        # Photo-related HTTP handlers
│   │   ├── albums.go        # Album-related HTTP handlers
│   │   └── validation.go    # Input validation
│   └── models/              # Data models and DTOs
├── frontend/
│   └── src/
│       ├── components/      # Vue components
│       ├── stores/         # Pinia state management
│       ├── api/            # API client
│       └── styles/         # CSS styling
└── config.example.yaml
```

## API Endpoints

### Photos
- `GET /api/photos/needsmetadata?album_id=X` - Photos without proper titles
- `GET /api/photos/:id` - Single photo details
- `PUT /api/photos/:id` - Update photo metadata

### Albums
- `GET /api/albums` - All normal albums
- `GET /api/albums/withphotocounts` - Albums containing photos needing titles

### Utility
- `GET /api/health` - Health check endpoint

## Technical Features

### Photo Title Detection
Advanced regex patterns identify photos needing titles:
- Generic 3-character camera prefixes: `^[A-Za-z0-9]{3}_[0-9]+(\\.\\w+)?$`
- Specific patterns for various camera types and formats
- UUID patterns with optional file extensions
- Cross-database compatibility (MySQL REGEXP, PostgreSQL ~, SQLite GLOB)

### Security
- Input validation with regex patterns
- SQL injection prevention using parameterized queries
- CORS middleware with security headers
- Query parameter validation and limits

### User Experience
- Album filtering (shows only albums with photos needing titles)
- Real-time photo removal after title updates
- Keyboard navigation (Cmd+J/K)
- Toast notifications in bottom-right corner
- Auto-focus on title field with text selection

## Deployment
Single binary deployment with embedded frontend assets. Simply copy the binary and config file to target system.