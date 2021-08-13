![tanuki](files/unminified/static/icon/favicon.ico) 
# Tanuki
Self hosted manga server + reader

## Features
- OPDS Support
- Multi-user support
- Supported formats: `.cbz`, `.zip`, `.cbr`, `.rar`
- Nested folders in library
- Track reading progress
- Thumbnail generation
- Single binary (~19.6 MB)
- Dark/light mode
- Metadata editor
- Mangadex downloader
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
scan_interval_minutes: 5
thumbnail_generation_interval_minutes: 60
max_uploaded_file_size_mib: 10
debug_mode: false
```
- To disable logging set `logging.log_to_file` and `logging.log_to_console` to `false`
- `logging.level` can be `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`
- `scan_interval_minutes` and `thumbnail_generation_interval_minutes` can be any non-negative integer. Set them to zero to disable the periodic tasks.
### Initial Login
On the first run Tanuki logs the default username and randomly generated password to STDOUT. It is advised to immediately change the password.

## OPDS
- The route for the OPDS catalog is `/opds/v1.2/catalog`
- This is the current [OPDS](https://specs.opds.io/) feature support:
### v1.2
- [x] Basic Auth
- [x] Viewing library
- [x] Downloading archive
- [x] Getting cover/thumbnail of archive
- [x] Search
- [x] Page streaming

## Changelog
### [0.9.3] - 2021-08-13
- Stops CORS policy error when searching for Mangadex entries

### [0.9.2] - 2021-08-13
- Removed delay on page load where navbar mobile shows until it detected it was on desktop

### [0.9.1] - 2021-08-13
- If unspecified intervals in config then uses defaults

### [0.9] - 2021-08-13
- Added favicon
- Added more documentation
- Docker image supports HTTP requests
- Docker image on the GitHub Container Registry
- Ignored correct files in .gitignore

### [0.8] - 2021-08-12
- Specifying a zero/negative interval disables periodic tasks
- Ability to specify config path with param
- Stops thumbnail generation blocking populating the DB
- Added OPDS Search & Page Streaming

### [0.7] - 2021-08-12
- Mangadex downloader

### [0.6] - 2021-08-06
- Webtoon support, in the reader modal you can select the webtoon option which removes gaps between rows
- Mobile UI overhaul, tanuki is now smoother and works better on mobile

### [0.5] - 2021-08-05
- Redundant CSS and JS code removed
- CSS classes all name appropriately
- CSS colours all specified in minify.go which makes it simpler to organise colours for both CSS themes
- Theme does not jitter on load anymore
- Alpine.js behaviour encapsulated into smaller classes which
  - enabled minification
  - reduces duplicity because behaviour is reproduced
- No more `can't find img.src` when loading images in the library view
- Many bug fixes, notable selecting options in the modal works correctly

### [0.4] - 2021-07-30
- Assets minified before embedding
- Data preloaded through templating engine to reduce API calls

### [0.3] - 2021-07-28
- Missing entries renamed to missing items
- Fixed bug where tanuki overwrites series/entry metadata with incorrect ones
- Tanuki now recognises missing series and entries as missing items
- Restructured project structure
  - Packages used by multiple main packages moved to the `/internal` directory
  - Packages used specifically to the user in the `/pkg` directory

### [0.2] - 2021-07-23
- Go routines speed up parsing series and adding them to the DB
- Properties like user progress bundles into one API call so less calls are made
- Series list is stored in order as catalog instead of being generated on the fly
- Metadata for each entry is stored in order instead of being generated on the fly
- Parsing/adding/generating series/thumbnails doesn't stop at the first error, it returns all errors which occurred at the end

### [0.1] - 2021-07-19
- Initial commit

## License
`MIT`