package models

// SizeVariantType represents the different size variants available in Lychee
type SizeVariantType int

const (
	// Based on Lychee's SizeVariantType enum:
	// https://github.com/LycheeOrg/Lychee/blob/master/app/Enum/SizeVariantType.php
	SizeVariantRaw         SizeVariantType = 0
	SizeVariantOriginal    SizeVariantType = 1
	SizeVariantMedium2x    SizeVariantType = 2
	SizeVariantMedium      SizeVariantType = 3
	SizeVariantSmall2x     SizeVariantType = 4
	SizeVariantSmall       SizeVariantType = 5
	SizeVariantThumb2x     SizeVariantType = 6
	SizeVariantThumb       SizeVariantType = 7
	SizeVariantPlaceholder SizeVariantType = 8
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

// PhotoWithSizeVariants extends PhotoWithAlbums to include size variants
type PhotoWithSizeVariants struct {
	PhotoWithAlbums
	ThumbnailPath *string `json:"thumbnail_path" db:"thumbnail_path"`
	LargePath     *string `json:"large_path" db:"large_path"`
	OriginalPath  *string `json:"original_path" db:"original_path"`
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
	case SizeVariantRaw:
		return "raw"
	case SizeVariantOriginal:
		return "original"
	case SizeVariantMedium2x:
		return "medium2x"
	case SizeVariantMedium:
		return "medium"
	case SizeVariantSmall2x:
		return "small2x"
	case SizeVariantSmall:
		return "small"
	case SizeVariantThumb2x:
		return "thumb2x"
	case SizeVariantThumb:
		return "thumb"
	case SizeVariantPlaceholder:
		return "placeholder"
	default:
		return "unknown"
	}
}