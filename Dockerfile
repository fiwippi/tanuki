FROM golang:1.16.3-alpine3.13 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/tanuki

FROM scratch

COPY --from=builder /app/bin/tanuki /tanuki
ENV DOCKER true
EXPOSE 8096
CMD ["/tanuki"]