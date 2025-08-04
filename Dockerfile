ARG BIN_NAME=lychee-meta-tool
ARG BIN_VERSION=<unknown>

# Build stage
FROM --platform=$BUILDPLATFORM node:22-alpine AS frontend-builder
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Go build stage
FROM --platform=$BUILDPLATFORM golang:1-alpine AS builder
ARG BIN_NAME
ARG BIN_VERSION
RUN apk add --no-cache gcc musl-dev sqlite-dev
RUN update-ca-certificates
WORKDIR /src/${BIN_NAME}
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /src/frontend/dist ./frontend/dist
RUN CGO_ENABLED=1 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME} .

# Final stage
FROM alpine:latest
ARG BIN_NAME
ARG BIN_VERSION
RUN apk add --no-cache ca-certificates sqlite
COPY --from=builder /src/${BIN_NAME}/out/${BIN_NAME} /usr/bin/${BIN_NAME}
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Create a non-root user
RUN addgroup -g 1000 lychee && \
    adduser -D -s /bin/sh -u 1000 -G lychee lychee

USER lychee
WORKDIR /home/lychee

EXPOSE 8080

ENTRYPOINT ["/usr/bin/lychee-meta-tool"]

LABEL license="LGPL3"
LABEL maintainer="Chris Dzombak <https://www.dzombak.com>"
LABEL org.opencontainers.image.authors="Chris Dzombak <https://www.dzombak.com>"
LABEL org.opencontainers.image.url="https://github.com/cdzombak/${BIN_NAME}"
LABEL org.opencontainers.image.documentation="https://github.com/cdzombak/${BIN_NAME}/blob/main/README.md"
LABEL org.opencontainers.image.source="https://github.com/cdzombak/${BIN_NAME}.git"
LABEL org.opencontainers.image.version="${BIN_VERSION}"
LABEL org.opencontainers.image.licenses="LGPL3"
LABEL org.opencontainers.image.title="${BIN_NAME}"
LABEL org.opencontainers.image.description="Quickly find & edit untitled photos in your Lychee photo library"