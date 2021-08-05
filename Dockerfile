FROM golang:1.16.3-alpine3.13 as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/minify minify.go && \
    bin/minify -input-dir files/unminified -output-dir files/minified && \
    rm -f bin/minify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/tanuki tanuki.go

FROM scratch

COPY --from=builder /app/bin/tanuki /tanuki
ENV DOCKER true
EXPOSE 8096
CMD ["/tanuki"]