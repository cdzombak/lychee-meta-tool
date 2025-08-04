package models

import "time"

type Album struct {
	ID          string     `json:"id" db:"id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	PublishedAt *time.Time `json:"published_at" db:"published_at"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description" db:"description"`
	OwnerID     int        `json:"owner_id" db:"owner_id"`
	IsNSFW      bool       `json:"is_nsfw" db:"is_nsfw"`
	IsPinned    bool       `json:"is_pinned" db:"is_pinned"`
	SortingCol  *string    `json:"sorting_col" db:"sorting_col"`
	SortingOrder *string   `json:"sorting_order" db:"sorting_order"`
	Copyright   *string    `json:"copyright" db:"copyright"`
	PhotoLayout *string    `json:"photo_layout" db:"photo_layout"`
	PhotoTimeline *string  `json:"photo_timeline" db:"photo_timeline"`
}

type AlbumResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type AlbumWithPhotoCount struct {
	Album
	PhotoCount int `json:"photo_count" db:"photo_count"`
}

type AlbumsResponse struct {
	Albums []AlbumResponse `json:"albums"`
}