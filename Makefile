build:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/tanuki main.go

run:
		./bin/tanuki

test:
		SERIES_PATH=$(SERIES_PATH) go test ./... -count=1

test-verbose:
		SERIES_PATH=$(SERIES_PATH) go test ./... -count=1 -v

clean:
		go clean
		rm -Rf bin
		rm -Rf data
		rm -Rf uploads
		rm -Rf config