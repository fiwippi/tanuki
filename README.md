![tanuki](files/unminified/static/icon/favicon.ico) 
# Tanuki
Self-hosted manga server + reader

## Features
- OPDS Support
- Multi-user support
- Supported formats: `.cbz`, `.zip`, `.cbr`, `.rar`
- Nested folders in library
- Track reading progress
- Thumbnail generation
- Single binary (~25.6 MB)
- Dark/light mode
- Mangadex downloader + subscriptions
- Responsive Desktop & Mobile UI
- Webtoon support (no row gaps)

## Installation
### Build from Source
```console
# Clone the repo
$ git clone https://github.com/fiwippi/tanuki.git

# Change the working directory to tanuki
$ cd tanuki

# Build the app and run it
$ make build && make run
```

### Docker
1. Clone the repository
2. Configure the `docker-compose.yml` file to set Tanuki to use the correct ports and mounted folders
3. Run `docker-compose up`
4. Open `localhost:8096` or another port if you specified one

#### GitHub Container Registry
An official container image exists at `ghcr.io/fiwippi/tanuki:latest`

## Usage
### CLI
```console
Usage of tanuki:
  -config string
        path to the config file, if it does not exist then it will be created (default "./config/config.yml")
  -recreate
        recreate the db on startup
```

### Config
Tanuki runs using a config which has a default path of `./config/config.yml` The config options and default values are specified below
```yaml
---
host: 0.0.0.0
port: "8096"
logging:
  level: info
  log_to_file: true
  log_to_console: true
paths:
  db: ./data/tanuki.db
  log: ./data/tanuki.log
  library: ./library
session_secret: tanuki-secret
scan_interval_minutes: 180            // How often Tanuki scans the library
subscriptions_interval_minutes: 1440  // How often Tanuki checks to see if new chapters have been released
max_uploaded_file_size_mib: 10
debug_mode: false
```
- To disable logging set `logging.log_to_file` and `logging.log_to_console` to `false`
- `logging.level` can be `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`
- `scan_interval_minutes` and `subscriptions_interval_minutes` can be any non-negative integer. To disable interval scanning set the value to `0`

⚠️ - Tanuki expects all archives in the library folder to be within their own folder. So, you can't have any standalone archives in the root of the library folder

### Initial Login
On the first run Tanuki logs the default username and randomly generated password to `STDOUT`. It is advised to immediately change the password

⚠️ - Tanuki generates all the thumbnails on startup which will cause a slight initial delay using it until thumbnail generation is done

### RAR Archives
If you supply tanuki with RAR archives (`.rar`, `.cbr`), their unarchive time to retrieve a single page is about 2 seconds compared to only milliseconds for a ZIP archives (`.zip`, `.cbz`) due to library constraints. For this reason consider converting all your files into ZIP archives

## OPDS
- The route for the OPDS catalog is `/opds/v1.2/catalog`
- This is the current [OPDS](https://specs.opds.io/) feature support:
### v1.2
- [x] Basic Auth
- [x] Viewing library
- [x] Downloading archive
- [x] Getting cover/thumbnail of archive
- [x] Page streaming
- [ ] Search

## License
`BSD-3-Clause`