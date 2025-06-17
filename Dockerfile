FROM --platform=$BUILDPLATFORM golang:1.24.3-bullseye AS build

WORKDIR /build

# Install cross-compilation tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETPLATFORM
ENV CGO_ENABLED=1

# Set up cross-compilation environment based on target platform
RUN case "${TARGETPLATFORM}" in \
    "linux/amd64") \
        CC=gcc CXX=g++ GOOS=linux GOARCH=amd64 make ;; \
    "linux/arm64") \
        CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ GOOS=linux GOARCH=arm64 make ;; \
    *) echo "Unsupported platform: ${TARGETPLATFORM}" && exit 1 ;; \
    esac

FROM debian:11-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates \
	&& rm -rf /var/lib/apt/lists/*

ARG TARGETPLATFORM

LABEL org.opencontainers.image.source=https://github.com/safedep/xbom
LABEL org.opencontainers.image.description="xbom is a tool to generate a Software Bill of Materials (SBOM) enriched with application metadata using static code analysis."
LABEL org.opencontainers.image.licenses=Apache-2.0

COPY --from=build /build/bin/xbom /usr/local/bin/xbom

ENTRYPOINT ["xbom"]
