# syntax=docker/dockerfile:1

ARG RUNTIME_IMAGE=docker.io/oven/bun:1-debian

FROM nixos/nix:latest AS builder

RUN printf 'experimental-features = nix-command flakes\nsandbox = false\nfilter-syscalls = false\n' \
    >> /etc/nix/nix.conf

WORKDIR /src
COPY . .

RUN nix build .#foundryctl --print-out-paths --no-link > /tmp/foundryctl-path

RUN mkdir -p /out \
 && cp "$(cat /tmp/foundryctl-path)/bin/foundryctl" /out/foundryctl

FROM ${RUNTIME_IMAGE} AS runtime

LABEL org.opencontainers.image.source="https://github.com/wrapped-owls/gontainer_foundryvtt"
LABEL org.opencontainers.image.description="FoundryVTT container runtime"

ENV FOUNDRY_DATA_PATH=/data \
    FOUNDRY_INSTALL_ROOT=/foundry \
    FOUNDRY_PATCH_MANIFEST=/etc/foundry/patches/manifest.yaml

RUN mkdir -p /data /foundry /etc/foundry/patches \
 && apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tzdata \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/foundryctl    /usr/local/bin/foundryctl
COPY patches/manifest.yaml             /etc/foundry/patches/manifest.yaml

VOLUME ["/data"]
EXPOSE 30000/tcp

HEALTHCHECK --start-period=3m --interval=30s --timeout=5s \
  CMD ["foundryctl", "healthcheck"]

ENTRYPOINT ["foundryctl"]
CMD ["run"]
