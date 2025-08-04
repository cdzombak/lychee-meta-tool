# Database Optimization Guide

This document provides recommendations for optimizing database performance for the Lychee Meta Tool.

## Recommended Database Indexes

The following indexes should be created to optimize the performance of queries used by the tool:

### MySQL/MariaDB

```sql
-- Optimize photo metadata queries
CREATE INDEX idx_photos_metadata ON photos(title, description, created_at);

-- Optimize photo ID lookups
CREATE INDEX idx_photos_id ON photos(id);

-- Optimize album relationship queries
CREATE INDEX idx_photos_album ON photos(old_album_id);

-- Optimize photo-album junction table
CREATE INDEX idx_photo_album_photo ON photo_album(photo_id);
CREATE INDEX idx_photo_album_album ON photo_album(album_id);

-- Optimize album queries
CREATE INDEX idx_albums_title ON base_albums(title);

-- Optimize tag album exclusion
CREATE INDEX idx_tag_albums_id ON tag_albums(id);
```

### PostgreSQL

```sql
-- Optimize photo metadata queries with regex support
CREATE INDEX idx_photos_metadata ON photos(title, description, created_at);

-- Optimize photo ID lookups
CREATE INDEX idx_photos_id ON photos(id);

-- Optimize album relationship queries
CREATE INDEX idx_photos_album ON photos(old_album_id);

-- Optimize photo-album junction table
CREATE INDEX idx_photo_album_photo ON photo_album(photo_id);
CREATE INDEX idx_photo_album_album ON photo_album(album_id);

-- Optimize album queries
CREATE INDEX idx_albums_title ON base_albums(title);

-- Optimize tag album exclusion
CREATE INDEX idx_tag_albums_id ON tag_albums(id);

-- PostgreSQL-specific: Enable regex optimization
CREATE INDEX idx_photos_title_pattern ON photos(title) WHERE title ~ '^(IMG_|DSC_?|DSCN|DSCF|CDZ_|P\d{7}|Screenshot)';
```

### SQLite

```sql
-- Optimize photo metadata queries
CREATE INDEX idx_photos_metadata ON photos(title, description, created_at);

-- Optimize photo ID lookups
CREATE INDEX idx_photos_id ON photos(id);

-- Optimize album relationship queries
CREATE INDEX idx_photos_album ON photos(old_album_id);

-- Optimize photo-album junction table
CREATE INDEX idx_photo_album_photo ON photo_album(photo_id);
CREATE INDEX idx_photo_album_album ON photo_album(album_id);

-- Optimize album queries
CREATE INDEX idx_albums_title ON base_albums(title);

-- Optimize tag album exclusion
CREATE INDEX idx_tag_albums_id ON tag_albums(id);
```

## Performance Optimization Tips

### 1. Query Optimization

- **Limit Result Sets**: The application enforces a maximum limit of 100 results per query
- **Use Offset Wisely**: For large datasets, consider cursor-based pagination instead of OFFSET
- **Monitor Slow Queries**: Enable slow query logging to identify performance bottlenecks

### 2. Database Configuration

#### MySQL/MariaDB
```ini
# my.cnf optimizations
innodb_buffer_pool_size = 256M  # Adjust based on available RAM
innodb_log_file_size = 64M
query_cache_size = 32M
max_connections = 100
```

#### PostgreSQL
```ini
# postgresql.conf optimizations
shared_buffers = 256MB  # Adjust based on available RAM
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB
max_connections = 100
```

#### SQLite
```sql
-- SQLite pragmas for better performance
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 10000;
PRAGMA temp_store = memory;
```

### 3. Regular Maintenance

#### MySQL/MariaDB
```sql
-- Analyze tables monthly
ANALYZE TABLE photos, base_albums, photo_album, tag_albums;

-- Optimize tables if needed
OPTIMIZE TABLE photos, base_albums, photo_album, tag_albums;
```

#### PostgreSQL
```sql
-- Update statistics weekly
ANALYZE photos, base_albums, photo_album, tag_albums;

-- Vacuum tables monthly
VACUUM ANALYZE photos, base_albums, photo_album, tag_albums;
```

#### SQLite
```sql
-- Analyze and optimize monthly
ANALYZE;
VACUUM;
REINDEX;
```

## Monitoring Queries

### Identify Slow Queries

#### MySQL/MariaDB
```sql
-- Enable slow query log
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 1;  -- Log queries taking > 1 second

-- Check slow queries
SELECT * FROM mysql.slow_log ORDER BY start_time DESC LIMIT 10;
```

#### PostgreSQL
```sql
-- Enable logging in postgresql.conf
log_min_duration_statement = 1000  -- Log queries taking > 1 second

-- Check current running queries
SELECT query, query_start, now() - query_start AS duration 
FROM pg_stat_activity 
WHERE state = 'active' 
ORDER BY duration DESC;
```

#### SQLite
```sql
-- Enable query plan analysis
EXPLAIN QUERY PLAN 
SELECT * FROM photos 
WHERE title REGEXP '^IMG_[0-9]+(\\.\\w+)?$' 
ORDER BY created_at DESC 
LIMIT 50;
```

## Expected Performance Metrics

With proper indexing, the application should achieve:

- **Photo metadata queries**: < 100ms for datasets up to 10,000 photos
- **Individual photo lookups**: < 10ms
- **Album listings**: < 50ms for up to 1,000 albums
- **Photo updates**: < 20ms per update

## Scaling Considerations

### Large Datasets (>100,000 photos)
- Consider partitioning the photos table by date
- Implement read replicas for query scaling
- Use connection pooling (implemented in the application)
- Consider caching frequently accessed data

### High Concurrency
- Monitor database connection usage
- Tune connection pool settings in the application
- Consider using a database proxy like PgBouncer (PostgreSQL) or ProxySQL (MySQL)

## Troubleshooting Performance Issues

1. **Slow Metadata Queries**:
   - Ensure indexes on title, description, and created_at exist
   - Check if regex patterns are being optimized by the database
   - Consider using full-text search for complex pattern matching

2. **High Memory Usage**:
   - Reduce connection pool size if necessary
   - Implement query result streaming for very large datasets
   - Monitor database buffer pool usage

3. **Lock Contention**:
   - Monitor for long-running transactions
   - Ensure photo updates are completing quickly
   - Consider using optimistic locking for photo updates

## Implementation Script

To apply the recommended indexes, run the appropriate SQL script for your database type. The application will automatically benefit from these optimizations without code changes.

For automated deployment, these index creation statements can be added to your database migration scripts.