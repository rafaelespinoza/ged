#
# Builder stage for auxiliary binaries
#
FROM alpine:3.24.1 AS builder_aux

LABEL com.rafaelespinoza.ged.stage_type='builder'

# TODO: get fzf v0.74.3
WORKDIR /src/fzf
RUN apk add --no-cache bash git && \
  git clone --depth 1 --branch v0.74.2 https://github.com/junegunn/fzf.git . && \
  ./install --bin && \
  mkdir -pv /outbin && \
  mv -v ./bin/fzf /outbin/fzf && \
  /outbin/fzf --version

WORKDIR /src/bkt
RUN wget https://github.com/dimo414/bkt/releases/download/0.8.2/bkt.v0.8.2.x86_64-unknown-linux-musl.zip && \
  gotzip=$(find . -name "bkt*.zip") && \
  ls -lah "${gotzip}" && \
  unzip "${gotzip}" && \
  chmod -c 755 ./bkt && \
  mv -v ./bkt /outbin/bkt && \
  /outbin/bkt --version

# TODO: get merman-cli from github releases once they release v0.8.0.
# This tool is a headless mermaid diagram maker.
# That version is expected to have a binary for aarch64.
# https://github.com/Latias94/merman/
# This also requires an update of github.com/dreampuf/mermaid.go to v0.4.0.

#
# Builder stage for main app binaries
#
FROM golang:alpine3.24 AS builder_main

LABEL com.rafaelespinoza.ged.stage_type='builder'
ENV CGO_ENABLED=0

WORKDIR /src/ged
COPY go.mod go.sum .
RUN go mod download && go mod verify
COPY . .
RUN apk add --no-cache git just && \
  just _MORE_LDFLAGS='-s' build && \
  mkdir -pv /outbin && \
  mv -v ./bin/main /outbin/ged && \
  /outbin/ged version && \
  cp -v ./explore-data.sh /outbin/

#
# Runtime stage
#
FROM alpine:3.24.1 AS runtime

ARG VERSION_BRANCH="unknown"
ARG VERSION_COMMIT="unknown"
ARG VERSION_TAG="dev"

# Make sure the base.name is consistent with the FROM above, and that it is
# the fully-qualified name (consistent with annotation spec).
# https://specs.opencontainers.org/image-spec/annotations/
LABEL org.opencontainers.image.base.name="docker.io/library/alpine:3.24.1" \
      org.opencontainers.image.title='ged-explore-data' \
      org.opencontainers.image.description='script for exploring genealogical data in GEDCOM format' \
      org.opencontainers.image.ref.name="${VERSION_BRANCH}" \
      org.opencontainers.image.revision="${VERSION_COMMIT}" \
      org.opencontainers.image.version="${VERSION_TAG}"

WORKDIR /data

COPY --from=builder_aux --chmod=755 /outbin/fzf /outbin/bkt /usr/local/bin
COPY --from=builder_main --chmod=644 /src/ged/testdata/*.ged /opt/testdata/
COPY --from=builder_main --chmod=755 /outbin/explore-data.sh /outbin/ged /usr/local/bin

# Add unprivileged user and remaining runtime deps. Ensure that the new
# user can run things.
RUN adduser -D ged && \
  apk add --no-cache bash jq && \
  su ged -c 'jq --version' && \
  su ged -c 'fzf --version' && \
  su ged -c 'bkt --version' && \
  su ged -c 'ged version' && \
  su ged -c 'explore-data.sh -h'

# The entrypoint script uses `fzf`, which will respond to SIGINT but not to
# the default stop signal SIGTERM. Do this to help the container stop
# immediately if doing `CONTAINER_TOOL stop`.
STOPSIGNAL SIGINT
USER ged
ENV GED_BIN=/usr/local/bin/ged

ENTRYPOINT ["/usr/local/bin/explore-data.sh"]
