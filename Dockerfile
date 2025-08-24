# -- Build tanuki
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o ./bin/tanuki cmd/tanuki/*.go
RUN CGO_ENABLED=0 go build -o ./bin/tanukictl cmd/tanukictl/*.go

# -- Run tanuki
FROM alpine:latest

COPY --from=builder /app/bin/tanuki /bin/tanuki
COPY --from=builder /app/bin/tanukictl /bin/tanukictl

ENTRYPOINT ["/bin/tanuki"]
