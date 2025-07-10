FROM golang:1.24-alpine AS builder

WORKDIR /src

# Leverage layer-caching: copy go.{mod,sum} first, then download deps
COPY go.mod ./
RUN go mod download

# Copy the rest of the source tree and compile a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/deepwiki


FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/deepwiki /usr/local/bin/deepwiki

WORKDIR /src

ENTRYPOINT ["deepwiki"]
