# Spec: lychee-meta-tool

This is a web-based tool for managing Lychee photo libraries. It allows the user to quickly set the title and description on photos that don't yet have them, and optionally move them to a different album.

## Functional Requirements

- Connects to the Lychee database (MySQL, PostgreSQL, or SQLite).
- Finds all photos that do not have a human-written title or description
    - This includes photos with empty strings or null values.
    - This includes photos whose names came from a digital camera or phone, possibly with a file extension (e.g. "IMG_1234", "CDZ_5678.jpg").
    - This includes photos whose titles are UUIDs, possibly with a file extension (e.g. "123e4567-e89b-12d3-a456-426614174000", "123e4567-e89b-12d3-a456-426614174000.jpg").
- Allows the user to set a title and description for each photo.
- Allows the user to move the photo to a different album.
- Allows the user to save changes to the photo back to the Lychee database.
- Provides a user interface to display the photos that need titles and descriptions.
- Provides a way to filter photos by album.
- Allows the user to navigate through the photos that need titles and descriptions.
- The program **does not** need to implement any authentication or authorization mechanisms; it assumes the user is already authenticated.

### Caveats

- Photos can only be added to albums that already exist in the Lychee database. The tool does not create new albums.
- Photos can only be added to "normal" albums, not tag albums or smart albums.

## UI Requirements

- The user interface displays a "filmstrip" of photos that need titles and descriptions along the top of the screen.
- The user can click on a photo in the filmstrip to view it in detail.
- The selected photo is displayed on the left half of the screen below the filmstrip; the right half of the screen contains input fields for the title and description.
- On selection of a new photo, the input fields should be populated with the current title and description of the photo. The title field is focused and its content is selected, allowing the user to quickly edit it.
- The user can hit Enter to save changes to the title while the Title field is focused.
- The user can hit Tab to move to the description field, and then hit Enter to save changes.
- The photo's current album is displayed, and the user can select a different album from an dropdown list.
- The dropdown list of albums is populated with all albums in the Lychee database, sorted alphabetically.
- The dropdown list of albums is searchable; it filters as the user types.
- After a photo is edited, and the user navigates to the next photo, the edited photo should not appear in the list of photos that need titles and descriptions.
- There are next/previous buttons on the left/right of the screen to navigate through the photos.
- The user can navigate through the photos using the keyboard shortcuts, inspired by Vim:
    - Command+J: Previous photo
    - Command+K: Next photo
- Toasts are displayed to indicate when a photo is saved successfully or if there is an error.
- The UI is tasteful and quick to respond, with no unnecessary delays.
- The UI should use the Medium photo size variant for the detail view, and the Thumb variant for the filmstrip.
- Images should be loaded lazily to improve performance, especially for large libraries.

## Technical Requirements

- The tool should be implemented as a web application using a modern JavaScript framework (e.g., React, Vue.js, or Angular).
- The backend should be implemented in Go, relying primarily on its standard library.
- Configuration should be given via a configuration file (in JSON or YAML).
- Configuration includes the Lychee site base URL, to be used to construct photo URLs.
- All frontend resources (HTML, CSS, JavaScript) should be embedded in the Go binary using `go:embed` to make deployment extremely simple as a single executable.
- The tool should be written in idiomatic and readable code. It should follow best practices for code organization and structure.
- Where feasible, unit test coverage is expected.

## Additional information

### Database Schema (MySQL)

This schema is provided to help the implementer understand the structure of the Lychee database. It is not exhaustive and may vary based on the Lychee version.

```sql
--- BEGIN TABLE base_albums
CREATE TABLE `base_albums` (
  `id` char(24) NOT NULL,
  `created_at` datetime(6) NOT NULL,
  `updated_at` datetime(6) NOT NULL,
  `published_at` datetime DEFAULT NULL,
  `title` varchar(100) NOT NULL,
  `description` text DEFAULT NULL,
  `owner_id` int(10) unsigned NOT NULL DEFAULT 0,
  `is_nsfw` tinyint(1) NOT NULL DEFAULT 0,
  `is_pinned` tinyint(1) NOT NULL DEFAULT 0,
  `sorting_col` varchar(30) DEFAULT NULL,
  `sorting_order` varchar(4) DEFAULT NULL,
  `copyright` varchar(300) DEFAULT NULL,
  `photo_layout` varchar(20) DEFAULT NULL,
  `photo_timeline` varchar(20) DEFAULT NULL,
  `_ai_description` text DEFAULT NULL,
  `_ai_description_ts` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `base_albums_owner_id_index` (`owner_id`),
  KEY `base_albums_published_at_index` (`published_at`),
  CONSTRAINT `base_albums_owner_id_foreign` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- END TABLE base_albums

-- BEGIN TABLE photos
CREATE TABLE `photos` (
  `id` char(24) NOT NULL,
  `created_at` datetime(6) NOT NULL,
  `updated_at` datetime(6) NOT NULL,
  `owner_id` int(10) unsigned NOT NULL DEFAULT 0,
  `old_album_id` char(24) DEFAULT NULL,
  `title` varchar(100) NOT NULL,
  `description` text DEFAULT NULL,
  `tags` text DEFAULT NULL,
  `license` varchar(20) NOT NULL DEFAULT 'none',
  `is_starred` tinyint(1) NOT NULL DEFAULT 0,
  `iso` varchar(255) DEFAULT NULL,
  `make` varchar(255) DEFAULT NULL,
  `model` varchar(255) DEFAULT NULL,
  `lens` varchar(255) DEFAULT NULL,
  `aperture` varchar(255) DEFAULT NULL,
  `shutter` varchar(255) DEFAULT NULL,
  `focal` varchar(255) DEFAULT NULL,
  `latitude` decimal(10,8) DEFAULT NULL,
  `longitude` decimal(11,8) DEFAULT NULL,
  `altitude` decimal(10,4) DEFAULT NULL,
  `img_direction` decimal(10,4) DEFAULT NULL,
  `location` varchar(255) DEFAULT NULL,
  `taken_at` datetime(6) DEFAULT NULL COMMENT 'relative to UTC',
  `taken_at_orig_tz` varchar(31) DEFAULT NULL COMMENT 'the timezone at which the photo has originally been taken',
  `initial_taken_at` datetime DEFAULT NULL COMMENT 'backup of the original taken_at value',
  `initial_taken_at_orig_tz` varchar(31) DEFAULT NULL COMMENT 'backup of the timezone at which the photo has originally been taken',
  `type` varchar(30) NOT NULL,
  `filesize` bigint(20) unsigned NOT NULL DEFAULT 0,
  `checksum` varchar(40) NOT NULL,
  `original_checksum` varchar(40) NOT NULL,
  `live_photo_short_path` varchar(255) DEFAULT NULL,
  `live_photo_content_id` varchar(255) DEFAULT NULL,
  `live_photo_checksum` varchar(40) DEFAULT NULL,
  `_ai_description` text DEFAULT NULL,
  `_ai_description_ts` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `photos_owner_id_foreign` (`owner_id`),
  KEY `photos_checksum_index` (`checksum`),
  KEY `photos_original_checksum_index` (`original_checksum`),
  KEY `photos_live_photo_content_id_index` (`live_photo_content_id`),
  KEY `photos_live_photo_checksum_index` (`live_photo_checksum`),
  KEY `photos_album_id_taken_at_index` (`old_album_id`,`taken_at`),
  KEY `photos_album_id_created_at_index` (`old_album_id`,`created_at`),
  KEY `photos_album_id_is_starred_index` (`old_album_id`,`is_starred`),
  KEY `photos_album_id_type_index` (`old_album_id`,`type`),
  KEY `photos_album_id_is_starred_created_at_index` (`old_album_id`,`is_starred`,`created_at`),
  KEY `photos_album_id_is_starred_taken_at_index` (`old_album_id`,`is_starred`,`taken_at`),
  KEY `photos_album_id_is_starred_type_index` (`old_album_id`,`is_starred`,`type`),
  KEY `photos_album_id_is_starred_title_index` (`old_album_id`,`is_starred`,`title`),
  KEY `photos_album_id_is_starred_description(128)_index` (`old_album_id`,`is_starred`,`description`(128)),
  CONSTRAINT `photos_owner_id_foreign` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- END TABLE photos

-- BEGIN TABLE photo_album
CREATE TABLE `photo_album` (
  `album_id` char(24) NOT NULL,
  `photo_id` char(24) NOT NULL,
  PRIMARY KEY (`photo_id`,`album_id`),
  KEY `photo_album_album_id_photo_id_index` (`album_id`,`photo_id`),
  KEY `photo_album_album_id_index` (`album_id`),
  KEY `photo_album_photo_id_index` (`photo_id`),
  CONSTRAINT `photo_album_album_id_foreign` FOREIGN KEY (`album_id`) REFERENCES `albums` (`id`),
  CONSTRAINT `photo_album_photo_id_foreign` FOREIGN KEY (`photo_id`) REFERENCES `photos` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- END TABLE photo_album

-- BEGIN TABLE tag_albums
CREATE TABLE `tag_albums` (
  `id` char(24) NOT NULL,
  `show_tags` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `tag_albums_id_foreign` FOREIGN KEY (`id`) REFERENCES `base_albums` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- END TABLE tag_albums

-- BEGIN TABLE size_variants
CREATE TABLE `size_variants` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `photo_id` char(24) NOT NULL,
  `type` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '0: original, ..., 6: thumb',
  `short_path` varchar(255) NOT NULL,
  `width` int(11) NOT NULL,
  `height` int(11) NOT NULL,
  `ratio` double NOT NULL DEFAULT 1,
  `filesize` bigint(20) unsigned NOT NULL DEFAULT 0,
  `storage_disk` varchar(255) NOT NULL DEFAULT 'images',
  PRIMARY KEY (`id`),
  UNIQUE KEY `size_variants_photo_id_type_unique` (`photo_id`,`type`),
  KEY `size_variants_short_path_index` (`short_path`),
  KEY `size_variants_photo_id_type_ratio_index` (`photo_id`,`type`,`ratio`),
  CONSTRAINT `size_variants_photo_id_foreign` FOREIGN KEY (`photo_id`) REFERENCES `photos` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=22545 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- END TABLE size_variants
```
