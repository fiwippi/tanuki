build:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/minify minify.go
		bin/minify -input-dir files/unminified -output-dir files/minified
		rm -f bin/minify

		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/tanuki tanuki.go

run:
		./bin/tanuki

clean:
		go clean
		rm -Rf bin
		rm -Rf data
		rm -Rf library
		rm -Rf config
		rm -Rf files/minified/static/css
		rm -Rf files/minified/static/js/api.js
		rm -Rf files/minified/static/js/auth.js
		rm -Rf files/minified/static/js/common.js
		rm -Rf files/minified/static/js/theme.js
		rm -Rf files/minified/templates
