# Test/lint/check image: pinned Go + git, fixture deps and build cache baked
# so every check container starts offline and warm (sqlite cold-compiles slowly).
FROM docker.io/library/golang:1.24-alpine@sha256:757779acac4af1b349a20f357c7296097b4a0b89da4ad0e370b339060077282a

RUN apk add --no-cache git
ENV GOTOOLCHAIN=local GOFLAGS=-modcacherw GOCACHE=/go/cache

COPY benchmarks/fixture/go.mod benchmarks/fixture/go.sum /warm/
# chmod: baked as root, used by any uid
RUN cd /warm && go mod download && go build modernc.org/sqlite \
	&& chmod -R a+rwX /go/pkg/mod /go/cache

RUN adduser -D -u 1000 runner
USER runner
WORKDIR /work
