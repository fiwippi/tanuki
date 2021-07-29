# Tanuki
Self hosted manga server + reader

## Features
- OPDS Support
- Multi-user support
- Supported formats: `.cbz`, `.zip`, `.cbr`, `.rar`
- Nested folders in library
- Track reading progress
- Thumbnail generation
- Single binary (19 MB)
- Dark/light mode
- Metadata editor

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
1. Configure the `docker-compose.yml` file to set Tanuki to use the correct ports and mounted folders
2. Run `docker-compose up`

## Usage
###
Tanuki runs using a config which has a default relative path of `./config/config.yml`
### Initial Login
On the first run Tanuki logs the default username and randomly generated password to STDOUT. It is advised to immediately change the password.

## License
`MIT`

## OPDS
This is the current [OPDS](https://specs.opds.io/) feature support
### v1.2
- [x] Basic Auth
- [x] Viewing library
- [x] Downloading archive
- [x] Getting cover/thumbnail of archive
- [ ] Search
- [ ] Page streaming
### v2.0
- N/A

## Development Roadmap
### Features
- Favicon
- Mangadex downloader
- Automatic download of new chapters
- Plugin support
- Specify config file path with param
- Docker file on the Github Container Registry
- Full OPDS feature support for v1.2 and v2.0

### Implementation Improvements
- WASM frontend for the API
- Clean up JS and CSS code/classes
- Make the UI more mobile friendly, e.g. so the modal loads in the centre of the phone screen
- Full well formatted documentation for Go + Javascript
- Single shareable modal class where esc key causes it to disappear
- Encapsulate retrieval and automatic generation of thumbnail if it doesn't exist into one function so that the API that the gin router acesses is simpler
- Clean up DB API by not returning an error for all functions since some don't need to
- Progress bar when uploading covers
- Use validators for binding
- Only include script src files where needed

## Changelog
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