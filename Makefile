build: minify
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/tanuki tanuki.go

run:
		./bin/tanuki

minify:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/minify minify.go
		bin/minify -input-dir files/unminified -output-dir files/minified
		rm -f bin/minify

clean:
		go clean
		rm -Rf bin
		rm -Rf data
		rm -Rf library
		rm -Rf config
		rm -Rf files/minified/static/css
		rm -Rf files/minified/static/icon
		rm -Rf files/minified/static/js/api.js
		rm -Rf files/minified/static/js/util.js
		rm -Rf files/minified/static/js/theme.js
		rm -Rf files/minified/static/js/mangadex.js
		rm -Rf files/minified/static/js/components
		rm -Rf files/minified/templates

test:
		go test ./... -v

test-quiet:
		go test ./...