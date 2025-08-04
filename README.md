# Lychee Meta Tool

Quickly find & edit untitled photos in your [Lychee](https://github.com/LycheeOrg/Lychee) photo library. Automatically identifies photos with generic camera names (`IMG_1234`, `CD5_5678`, etc.) and provides an efficient interface for adding meaningful titles.

## Features

- **Smart Detection**: Automatically finds photos with generic camera names, UUIDs, or empty titles
- **Album Filtering**: Work on photos from specific albums only
- **Keyboard Navigation**: Ctrl+J/K for previous/next photo
- **Single Binary Deployment**: All frontend assets embedded
- **Multi-Database Support**: MySQL, PostgreSQL, SQLite
- **AI Title Suggestions:** optional Ollama integration for title suggestions

### Photo Detection

Identifies photos needing titles by looking for the following patterns:

- 3-character camera prefixes: `CD5_1234`, `IMG_5678`, `DSZ_9012`
- UUID-based filenames
- Screenshot and timestamp patterns
- Empty or null titles

## Installation

### macOS via Homebrew

```shell
brew install cdzombak/oss/lychee-meta-tool
```

### Debian via apt repository

[Install my Debian repository](https://www.dzombak.com/blog/2025/06/updated-instructions-for-installing-my-debian-package-repositories/) if you haven't already:

```shell
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://dist.cdzombak.net/keys/dist-cdzombak-net.gpg -o /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo chmod 644 /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo mkdir -p /etc/apt/sources.list.d
sudo curl -fsSL https://dist.cdzombak.net/cdzombak-oss.sources -o /etc/apt/sources.list.d/cdzombak-oss.sources
sudo chmod 644 /etc/apt/sources.list.d/cdzombak-oss.sources
sudo apt update
```

Then install `lychee-meta-tool` via `apt-get`:

```shell
sudo apt-get install lychee-meta-tool
```

### Manual installation from build artifacts

Pre-built binaries for Linux and macOS on various architectures are downloadable from each [GitHub Release](https://github.com/cdzombak/lychee-meta-tool/releases). Debian packages for each release are available as well.

### Build and install locally

```shell
git clone https://github.com/cdzombak/lychee-meta-tool.git
cd lychee-meta-tool
make build

cp out/lychee-meta-tool $INSTALL_DIR
```

## Docker images

Docker images are available for a variety of Linux architectures from [Docker Hub](https://hub.docker.com/r/cdzombak/lychee-meta-tool) and [GHCR](https://github.com/cdzombak/dirshard/pkgs/container/lychee-meta-tool). Images are based on the `scratch` image and are as small as possible.

Run them via, for example:

```shell
docker run --rm cdzombak/lychee-meta-tool:1 [OPTIONS]
docker run --rm ghcr.io/cdzombak/lychee-meta-tool:1 [OPTIONS]
```

## Configuration

Configuration is provided via a JSON or YAML file. See [`config.example.yaml`](config.example.yaml). 

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
4. Press Enter to save and move to the next photo
5. Use Ctrl+J/K for keyboard navigation
6. _(optional)_ Use Ctrl-I for AI title suggestion

## License

MIT License; see [`LICENSE`](LICENSE) in this repo.

## Author

[Claude Code](https://www.anthropic.com/claude-code) wrote this code with management and changes by Chris Dzombak ([dzombak.com](https://www.dzombak.com) / [github.com/cdzombak](https://www.github.com/cdzombak)).
