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

## Changelog
### [0.1] - 2021-07-19
Initial commit

## Development Roadmap
### Features
- Favicon
- Mangadex downloader
- More supported archive formats, e.g. `.tar`
- More supported image formats e.g. `.avif`
- Automatic download of new chapters
- Plugin support
- Specify config file path with param
- Docker file on the Github Container Registry
- Full OPDS feature support for v1.2 and v2.0

### Implementation Improvements
- WASM frontend for the API
- Minimise final JS and CSS file
- Use more go routines to speed up data processing functions
- Make the UI more mobile friendly, e.g. so the modal loads in the centre of the phone screen
- Reduce API calls by:
    - serving some data through the golang templating engine, e.g. sid, eid or isAdmin
    - calling the api for multiple properties at once instead of each property individually, e.g. call for progress for each entry of the series in once call vs separate call for each entry
- Full well formatted documentation for Go + Javascript
- Single shareable modal class where esc key causes it to disappear
- Try and remove some dependencies e.g. xid to reduce file size
- Log more less important routes as trace
- Encapsulate retrievel and autoamtic generation of thumbnail if it doesn't exist into one function so that the API that the gin router acesses is simpler
- Instead of sorting SeriesList and SeriesEntries, they should already be stored in a sorted order
- Can the function which retrieves Series Entries just us `keySeriesData` to make data retrieval simpler
- Generate thumbnails shouldn't stop at the first error, it should continue to the end and then return an error for each thumbnail it failed to generate
- Clean up DB API by not returning an error for all functions since some don't need to
- Store the user progress data in the specific entry bucket for each so if a series is deleted/missing/changed it is easier to update the user's data
- Store series list in the DB so no need to generate it on the fly each time
- Progress bar when uploading covers

## Tests
To run tests, an example archive file has to be supplied using the `SERIES_PATH` parameter, so for example:
```console
make test SERIES_PATH=./series/chapter.cbz
```