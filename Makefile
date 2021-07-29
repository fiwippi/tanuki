build:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/tanuki main.go

run:
		./bin/tanuki

clean:
		go clean
		rm -Rf bin
		rm -Rf data
		rm -Rf library
		rm -Rf config