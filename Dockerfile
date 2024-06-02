# -- Build tanuki
FROM golang:1.22-alpine as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o ./bin/tanuki cmd/*.go

# -- Run tanuki
FROM alpine:latest

COPY --from=builder /app/bin/tanuki /bin/tanuki

ENTRYPOINT ["/bin/tanuki"]