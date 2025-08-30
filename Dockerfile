# -- Build tanuki
FROM golang:1.24-alpine AS builder

ARG VERSION

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-X github.com/fiwippi/tanuki/v2.Version=$VERSION" -o ./bin/tanuki cmd/tanuki/*.go
RUN CGO_ENABLED=0 go build -ldflags "-X github.com/fiwippi/tanuki/v2.Version=$VERSION" -o ./bin/tanukictl cmd/tanukictl/*.go

# -- Run tanuki
FROM alpine:latest

COPY --from=builder /app/bin/tanuki /bin/tanuki
COPY --from=builder /app/bin/tanukictl /bin/tanukictl

ENTRYPOINT ["/bin/tanuki"]
