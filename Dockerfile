# Build stage
FROM golang:1.25.7-alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /crontask ./cmd/crontask/main.go

# Get certs with configured mirrors
FROM alpine:3.23 AS certs

# Configure Alpine package mirrors (choose one option below)

# OPTION 1: Use main Alpine CDN (fastest worldwide)
RUN echo "http://mirror-linux.runflare.com/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror-linux.runflare.com/alpine/v3.23/community" >> /etc/apk/repositories

# OPTION 2: Use multiple mirrors (redundancy)
# RUN echo -e "http://dl-cdn.alpinelinux.org/alpine/v3.23/main\n\
# http://mirror.math.princeton.edu/pub/alpinelinux/v3.23/main\n\
# http://mirrors.ocf.berkeley.edu/alpine/v3.23/main" > /etc/apk/repositories && \
# echo -e "http://dl-cdn.alpinelinux.org/alpine/v3.23/community\n\
# http://mirror.math.princeton.edu/pub/alpinelinux/v3.23/community\n\
# http://mirrors.ocf.berkeley.edu/alpine/v3.23/community" >> /etc/apk/repositories

# OPTION 3: Add non-free repositories if needed
# RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.23/main" > /etc/apk/repositories && \
#     echo "http://dl-cdn.alpinelinux.org/alpine/v3.23/community" >> /etc/apk/repositories && \
#     echo "http://dl-cdn.alpinelinux.org/alpine/v3.23/non-free" >> /etc/apk/repositories

RUN apk update && apk --no-cache add ca-certificates

# Final stage
FROM busybox:glibc

# Copy certificates
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
COPY --from=builder /crontask .

RUN adduser -D -u 1000 appuser && \
    chown appuser:appuser /app/crontask
USER appuser

CMD ["/app/crontask"]
