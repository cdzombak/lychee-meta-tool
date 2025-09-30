package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cdzombak/lychee-meta-tool/backend/models"
)

func (db *DB) GetPhotosNeedingMetadata(albumID *string, limit, offset int) ([]models.PhotoWithSizeVariants, error) {
	query := `
		SELECT
			p.id, p.created_at, p.updated_at, p.owner_id, p.old_album_id,
			p.title, p.description, p.license, p.is_starred,
			p.iso, p.make, p.model, p.lens, p.aperture, p.shutter, p.focal,
			p.latitude, p.longitude, p.altitude, p.img_direction, p.location,
			p.taken_at, p.type, p.filesize, p.checksum,
			a.title as album_title,
			sv_thumb.short_path as thumbnail_path,
			sv_large.short_path as large_path,
			sv_original.short_path as original_path
		FROM photos p
		LEFT JOIN base_albums a ON p.old_album_id = a.id
		LEFT JOIN size_variants sv_thumb ON p.id = sv_thumb.photo_id AND sv_thumb.type = 6
		LEFT JOIN size_variants sv_large ON p.id = sv_large.photo_id AND sv_large.type = 3
		LEFT JOIN size_variants sv_original ON p.id = sv_original.photo_id AND sv_original.type = 0
		WHERE (
			p.title = '' OR p.title IS NULL OR
			p.title REGEXP '^[A-Za-z0-9]{3}_[0-9]+(\\.\\w+)?$' OR
			p.title REGEXP '^P[0-9]{7}(\\.\\w+)?$' OR
			p.title REGEXP '^[0-9]{8}_[0-9]{6}(\\.\\w+)?$' OR
			p.title REGEXP '^IMG-[0-9]{8}-WA[0-9]{4}(\\.\\w+)?$' OR
			p.title REGEXP '^Screenshot.*(\\.\\w+)?$' OR
			p.title REGEXP '^[0-9a-fA-F]{8}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{12}(\\.\\w+)?$'
		)`

	args := []interface{}{}
	
	if albumID != nil {
		query += " AND p.old_album_id = ?"
		args = append(args, *albumID)
	}

	query += " ORDER BY p.created_at DESC"
	
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
		
		if offset > 0 {
			query += " OFFSET ?"
			args = append(args, offset)
		}
	}

	// Adjust query for PostgreSQL if needed
	switch db.driver {
	case "postgres":
		query = db.convertToPostgreSQL(query)
	case "sqlite":
		query = db.convertToSQLite(query)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query photos: %w", err)
	}
	defer rows.Close()

	var photos []models.PhotoWithSizeVariants
	for rows.Next() {
		var photo models.PhotoWithSizeVariants
		err := rows.Scan(
			&photo.ID, &photo.CreatedAt, &photo.UpdatedAt, &photo.OwnerID, &photo.AlbumID,
			&photo.Title, &photo.Description, &photo.License, &photo.IsStarred,
			&photo.ISO, &photo.Make, &photo.Model, &photo.Lens, &photo.Aperture, &photo.Shutter, &photo.Focal,
			&photo.Latitude, &photo.Longitude, &photo.Altitude, &photo.ImgDirection, &photo.Location,
			&photo.TakenAt, &photo.Type, &photo.Filesize, &photo.Checksum,
			&photo.AlbumTitle, &photo.ThumbnailPath, &photo.LargePath, &photo.OriginalPath,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan photo: %w", err)
		}
		photos = append(photos, photo)
	}

	return photos, nil
}

func (db *DB) GetPhotoByID(id string) (*models.PhotoWithSizeVariants, error) {
	query := `
		SELECT
			p.id, p.created_at, p.updated_at, p.owner_id, p.old_album_id,
			p.title, p.description, p.license, p.is_starred,
			p.iso, p.make, p.model, p.lens, p.aperture, p.shutter, p.focal,
			p.latitude, p.longitude, p.altitude, p.img_direction, p.location,
			p.taken_at, p.type, p.filesize, p.checksum,
			a.title as album_title,
			sv_thumb.short_path as thumbnail_path,
			sv_large.short_path as large_path,
			sv_original.short_path as original_path
		FROM photos p
		LEFT JOIN base_albums a ON p.old_album_id = a.id
		LEFT JOIN size_variants sv_thumb ON p.id = sv_thumb.photo_id AND sv_thumb.type = 6
		LEFT JOIN size_variants sv_large ON p.id = sv_large.photo_id AND sv_large.type = 3
		LEFT JOIN size_variants sv_original ON p.id = sv_original.photo_id AND sv_original.type = 0
		WHERE p.id = ?`

	var photo models.PhotoWithSizeVariants
	err := db.QueryRow(query, id).Scan(
		&photo.ID, &photo.CreatedAt, &photo.UpdatedAt, &photo.OwnerID, &photo.AlbumID,
		&photo.Title, &photo.Description, &photo.License, &photo.IsStarred,
		&photo.ISO, &photo.Make, &photo.Model, &photo.Lens, &photo.Aperture, &photo.Shutter, &photo.Focal,
		&photo.Latitude, &photo.Longitude, &photo.Altitude, &photo.ImgDirection, &photo.Location,
		&photo.TakenAt, &photo.Type, &photo.Filesize, &photo.Checksum,
		&photo.AlbumTitle, &photo.ThumbnailPath, &photo.LargePath, &photo.OriginalPath,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get photo: %w", err)
	}

	return &photo, nil
}

func (db *DB) UpdatePhoto(id string, update models.PhotoUpdate) error {
	// Build update query with explicit field handling to prevent SQL injection
	var query string
	var args []interface{}
	
	// Determine which fields to update
	updateTitle := update.Title != nil
	updateDescription := update.Description != nil
	
	if !updateTitle && !updateDescription {
		// No photo metadata to update, just handle album change if needed
		if update.AlbumID != nil {
			if err := db.UpdatePhotoAlbum(id, *update.AlbumID); err != nil {
				return fmt.Errorf("failed to update photo album: %w", err)
			}
		}
		return nil
	}
	
	// Build query with explicit field combinations to avoid string concatenation
	if updateTitle && updateDescription {
		query = "UPDATE photos SET title = ?, description = ?, updated_at = NOW() WHERE id = ?"
		args = []interface{}{*update.Title, *update.Description, id}
	} else if updateTitle {
		query = "UPDATE photos SET title = ?, updated_at = NOW() WHERE id = ?"
		args = []interface{}{*update.Title, id}
	} else if updateDescription {
		query = "UPDATE photos SET description = ?, updated_at = NOW() WHERE id = ?"
		args = []interface{}{*update.Description, id}
	}

	// Adjust for SQLite's datetime function
	if db.driver == "sqlite" {
		query = strings.Replace(query, "NOW()", "datetime('now')", 1)
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update photo: %w", err)
	}

	// Handle album change separately
	if update.AlbumID != nil {
		if err := db.UpdatePhotoAlbum(id, *update.AlbumID); err != nil {
			return fmt.Errorf("failed to update photo album: %w", err)
		}
	}

	return nil
}

func (db *DB) UpdatePhotoAlbum(photoID, albumID string) error {
	// First update the old_album_id in photos table
	query := "UPDATE photos SET old_album_id = ?, updated_at = NOW() WHERE id = ?"
	args := []interface{}{albumID, photoID}

	if db.driver == "sqlite" {
		query = strings.Replace(query, "NOW()", "datetime('now')", 1)
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update photo album_id: %w", err)
	}

	// Remove existing photo_album relationships
	_, err = db.Exec("DELETE FROM photo_album WHERE photo_id = ?", photoID)
	if err != nil {
		return fmt.Errorf("failed to delete old photo_album relationships: %w", err)
	}

	// Add new photo_album relationship
	_, err = db.Exec("INSERT INTO photo_album (photo_id, album_id) VALUES (?, ?)", photoID, albumID)
	if err != nil {
		return fmt.Errorf("failed to insert new photo_album relationship: %w", err)
	}

	return nil
}

func (db *DB) GetAlbums() ([]models.Album, error) {
	query := `
		SELECT 
			id, created_at, updated_at, published_at, title, description,
			owner_id, is_nsfw, is_pinned, sorting_col, sorting_order,
			copyright, photo_layout, photo_timeline
		FROM base_albums 
		WHERE id NOT IN (SELECT id FROM tag_albums)
		ORDER BY title ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query albums: %w", err)
	}
	defer rows.Close()

	var albums []models.Album
	for rows.Next() {
		var album models.Album
		err := rows.Scan(
			&album.ID, &album.CreatedAt, &album.UpdatedAt, &album.PublishedAt,
			&album.Title, &album.Description, &album.OwnerID, &album.IsNSFW,
			&album.IsPinned, &album.SortingCol, &album.SortingOrder,
			&album.Copyright, &album.PhotoLayout, &album.PhotoTimeline,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan album: %w", err)
		}
		albums = append(albums, album)
	}

	return albums, nil
}

func (db *DB) GetAlbumsWithPhotoCounts() ([]models.AlbumWithPhotoCount, error) {
	query := `
		SELECT 
			a.id, a.created_at, a.updated_at, a.published_at, a.title, a.description,
			a.owner_id, a.is_nsfw, a.is_pinned, a.sorting_col, a.sorting_order,
			a.copyright, a.photo_layout, a.photo_timeline,
			COUNT(p.id) as photo_count
		FROM base_albums a
		LEFT JOIN photos p ON a.id = p.old_album_id AND (
			p.title = '' OR p.title IS NULL OR
			p.title REGEXP '^[A-Za-z0-9]{3}_[0-9]+(\\.\\w+)?$' OR
			p.title REGEXP '^P[0-9]{7}(\\.\\w+)?$' OR
			p.title REGEXP '^[0-9]{8}_[0-9]{6}(\\.\\w+)?$' OR
			p.title REGEXP '^IMG-[0-9]{8}-WA[0-9]{4}(\\.\\w+)?$' OR
			p.title REGEXP '^Screenshot.*(\\.\\w+)?$' OR
			p.title REGEXP '^[0-9a-fA-F]{8}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{12}(\\.\\w+)?$'
		)
		WHERE a.id NOT IN (SELECT id FROM tag_albums)
		GROUP BY a.id, a.created_at, a.updated_at, a.published_at, a.title, a.description,
				 a.owner_id, a.is_nsfw, a.is_pinned, a.sorting_col, a.sorting_order,
				 a.copyright, a.photo_layout, a.photo_timeline
		HAVING COUNT(p.id) > 0
		ORDER BY a.title ASC`

	// Adjust query for different databases
	switch db.driver {
	case "postgres":
		query = db.convertToPostgreSQL(query)
	case "sqlite":
		query = db.convertToSQLiteWithPhotoCounts(query)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query albums with photo counts: %w", err)
	}
	defer rows.Close()

	var albums []models.AlbumWithPhotoCount
	for rows.Next() {
		var album models.AlbumWithPhotoCount
		err := rows.Scan(
			&album.ID, &album.CreatedAt, &album.UpdatedAt, &album.PublishedAt,
			&album.Title, &album.Description, &album.OwnerID, &album.IsNSFW,
			&album.IsPinned, &album.SortingCol, &album.SortingOrder,
			&album.Copyright, &album.PhotoLayout, &album.PhotoTimeline,
			&album.PhotoCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan album with photo count: %w", err)
		}
		albums = append(albums, album)
	}

	return albums, nil
}

func (db *DB) convertToPostgreSQL(query string) string {
	// Convert MySQL REGEXP to PostgreSQL ~
	query = strings.ReplaceAll(query, "REGEXP", "~")
	// Convert MySQL backticks to PostgreSQL double quotes (if any)
	query = strings.ReplaceAll(query, "`", "\"")
	return query
}

func (db *DB) convertToSQLite(query string) string {
	// SQLite doesn't support REGEXP by default, we'll use LIKE patterns instead
	// This is a simplified conversion - in production, you might want to enable REGEXP extension
	query = strings.ReplaceAll(query, "p.title REGEXP '^[A-Za-z0-9]{3}_[0-9]+(\\.\\w+)?$'", "(p.title GLOB '???_*' AND LENGTH(p.title) >= 5)")
	query = strings.ReplaceAll(query, "p.title REGEXP '^P[0-9]{7}(\\.\\w+)?$'", "p.title GLOB 'P*'")
	query = strings.ReplaceAll(query, "p.title REGEXP '^[0-9]{8}_[0-9]{6}(\\.\\w+)?$'", "p.title GLOB '*_*'")
	query = strings.ReplaceAll(query, "p.title REGEXP '^IMG-[0-9]{8}-WA[0-9]{4}(\\.\\w+)?$'", "p.title GLOB 'IMG-*-WA*'")
	query = strings.ReplaceAll(query, "p.title REGEXP '^Screenshot.*(\\.\\w+)?$'", "p.title GLOB 'Screenshot*'")
	// UUID pattern is complex, we'll use a simpler check
	query = strings.ReplaceAll(query, "p.title REGEXP '^[0-9a-fA-F]{8}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{12}(\\.\\w+)?$'", "(LENGTH(p.title) = 32 OR LENGTH(p.title) = 36)")
	return query
}

func (db *DB) convertToSQLiteWithPhotoCounts(query string) string {
	// SQLite doesn't support REGEXP by default, we'll use LIKE patterns instead
	// This is a simplified conversion - in production, you might want to enable REGEXP extension
	query = strings.ReplaceAll(query, "p.title REGEXP '^[A-Za-z0-9]{3}_[0-9]+(\\.\\w+)?$'", "(p.title GLOB '???_*' AND LENGTH(p.title) >= 5)")
	query = strings.ReplaceAll(query, "p.title REGEXP '^P[0-9]{7}(\\.\\w+)?$'", "p.title GLOB 'P*'")
	query = strings.ReplaceAll(query, "p.title REGEXP '^[0-9]{8}_[0-9]{6}(\\.\\w+)?$'", "p.title GLOB '*_*'")
	query = strings.ReplaceAll(query, "p.title REGEXP '^IMG-[0-9]{8}-WA[0-9]{4}(\\.\\w+)?$'", "p.title GLOB 'IMG-*-WA*'")
	query = strings.ReplaceAll(query, "p.title REGEXP '^Screenshot.*(\\.\\w+)?$'", "p.title GLOB 'Screenshot*'")
	// UUID pattern is complex, we'll use a simpler check
	query = strings.ReplaceAll(query, "p.title REGEXP '^[0-9a-fA-F]{8}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{12}(\\.\\w+)?$'", "(LENGTH(p.title) = 32 OR LENGTH(p.title) = 36)")
	return query
}