# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
RUN apk add --no-cache git ca-certificates make bash
WORKDIR /src

COPY . .

ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BRANCH="unknown"

# Build via Makefile (sets correct ldflags)
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
RUN make go-install

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -H -s /sbin/nologin telegraf

COPY --from=build /go/bin/telegraf /usr/bin/telegraf

# Most Telegraf images run telegraf directly; config is mounted at runtime
USER telegraf
ENTRYPOINT ["telegraf"]
