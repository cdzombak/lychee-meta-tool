package models

// SizeVariantType represents the different size variants available in Lychee
type SizeVariantType int

const (
	// Based on Lychee's size variant types
	SizeVariantOriginal    SizeVariantType = 0
	SizeVariantSmall2x     SizeVariantType = 1
	SizeVariantSmall       SizeVariantType = 2
	SizeVariantMedium2x    SizeVariantType = 3
	SizeVariantMedium      SizeVariantType = 4
	SizeVariantSmallThumb  SizeVariantType = 5
	SizeVariantThumb       SizeVariantType = 6
)

// SizeVariant represents a photo size variant in the Lychee database
type SizeVariant struct {
	ID          int64           `json:"id" db:"id"`
	PhotoID     string          `json:"photo_id" db:"photo_id"`
	Type        SizeVariantType `json:"type" db:"type"`
	ShortPath   string          `json:"short_path" db:"short_path"`
	Width       int             `json:"width" db:"width"`
	Height      int             `json:"height" db:"height"`
	Ratio       float64         `json:"ratio" db:"ratio"`
	Filesize    int64           `json:"filesize" db:"filesize"`
	StorageDisk string          `json:"storage_disk" db:"storage_disk"`
}

// PhotoWithSizeVariants extends PhotoWithAlbum to include size variants
type PhotoWithSizeVariants struct {
	PhotoWithAlbum
	ThumbnailPath *string `json:"thumbnail_path" db:"thumbnail_path"`
	OriginalPath  *string `json:"original_path" db:"large_path"`
}

// GetThumbnailVariant returns the thumbnail size variant type
func GetThumbnailVariant() SizeVariantType {
	return SizeVariantThumb
}

// GetOriginalVariant returns the original size variant type for detail view
func GetOriginalVariant() SizeVariantType {
	return SizeVariantOriginal
}

// String returns a string representation of the size variant type
func (s SizeVariantType) String() string {
	switch s {
	case SizeVariantOriginal:
		return "original"
	case SizeVariantSmall2x:
		return "small2x"
	case SizeVariantSmall:
		return "small"
	case SizeVariantMedium2x:
		return "medium2x"
	case SizeVariantMedium:
		return "medium"
	case SizeVariantSmallThumb:
		return "small_thumb"
	case SizeVariantThumb:
		return "thumb"
	default:
		return "unknown"
	}
}