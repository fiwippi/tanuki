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
- Single binary (~19.1 MB)
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
scan_interval_minutes: 180
max_uploaded_file_size_mib: 10
debug_mode: false
```
- To disable logging set `logging.log_to_file` and `logging.log_to_console` to `false`
- `logging.level` can be `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`
- `scan_interval_minutes`  can be any non-negative integer. To disable interval scanning set the value to `0`;
### Initial Login
On the first run Tanuki logs the default username and randomly generated password to `STDOUT`. It is advised to immediately change the password

⚠️ - Tanuki generates all the thumbnails on startup which will cause a slight initial delay using it until thumbnail generation is done
### Rar Archives
If you supply tanuki with Rar archives (`.rar`, `.cbr`), their unarchive time to retrieve a single page is about 2 seconds compared to only milliseconds for a Zip archives (`.zip`, `.cbz`) due to library constraints. For this reason consider converting all your files into Zip archives

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
### [0.16.2] - 2022-02-18
- Support Mangadex download with no chapter and volume

### [0.16.1] - 2022-02-17
- Fixed bugs

### [0.16] - 2022-02-15
- Tanuki automatically scans for new downloads once all downloads are done

### [0.15] - 2022-02-15
⚠️ - Breaking changes are made to the database so you should recreate yours!
- Fixed bugs when tracking user progress
- Session token refreshed regardless of time left
- More informative errors and info added to the backend

### [0.14] - 2022-02-07
- More responsive UI
- Removed unused internal packages
- License changed to `BSD-3-Clause`
- Default session cookie life changed to 3 days
- Updated to new Mangadex API change for downloading chapters
- Archive page sorting interprets underscores before other characters
- Remembers last page user tried to access if they're not authed anymore and redirects there after subsequent login

### [0.13.1] - 2021-11-03
- Removed trailing dots from names of downloaded manga
- Default scan interval changed to 3 hours

### [0.13] - 2021-11-03
- Stopped program consuming exhaustive amount of file descriptors
- Filepath sanitiser does not remove apostrophes and exclamation marks from downloaded Mangadex titles
- Autofocus for searchbar for Mangadex
- No autocorrect and autocapitalisation for the login screen
- Minor documentation changes

### [0.12.3] - 2021-09-27
- Fixed page ordering bug introduced in 0.12.1

### [0.12.2] - 2021-09-26
- More edge cases handled when decoding images

### [0.12.1] - 2021-09-23
- Fixed ordering for some pages

### [0.12] - 2021-09-23
- Better support for unicode characters
- Correct page displayed instead of page with similar name
- Loading animation for viewing the DB + searching and downloading from Mangadex
- Updated Mangadex functionality broken by Mangadex API change
- Series progress calculation takes into account all entries instead of only ones which have been accessed

### [0.11] - 2021-08-24
- Thumbnail generation generated lazily instead of with a recurring job, this stops the throttling of the db whenever the job was running and speeds up page loading
- Mangas with pages specified only by numbers (with no text) no longer sorted in inverse order
- Stopped invalid progress displaying as NaN client-side
- Natural ordering improved so entries in a series now assigned correct order (until I manage to break it again)
- Invalid API requests automatically redirect the user to the login page as opposed to only when accessing the frontend routes
- When setting progress for a whole series, entry progress is created if it is previously nil
- Loading animation when scanning in the db or manually generating thumbnails

### [0.10] - 2021-08-17
- Non-english titles displayed for Mangadex entries if english ones don't exist
- Mangadex archive pages zero padded in the archive
- Archives walked in proper lexical order (but lowercase first)
- Entry progress only created when reading the entry instead of when accessing the series
- Default entry progress is displayed as 0.00% instead of N/A
- Not all thumbnails have to be loaded to display them client side
- Series in the catalog and entries in each series are ordered naturally

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
`BSD-3-Clause`