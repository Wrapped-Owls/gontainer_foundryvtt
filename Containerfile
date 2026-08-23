# syntax=docker/dockerfile:1

FROM nixos/nix:2.35.2 AS builder

RUN printf '%s\n' 'experimental-features = nix-command flakes' 'sandbox = false' \
      'filter-syscalls = false' >> /etc/nix/nix.conf

WORKDIR /src

COPY flake.nix flake.lock ./
COPY nix/ ./nix/

FROM builder AS bot-build

COPY . .
RUN nix build .#taverncord --print-out-paths --no-link > /tmp/bot-path \
 && mkdir -p /out \
 && cp "$(cat /tmp/bot-path)/bin/taverncord" /out/taverncord

FROM alpine:3.21 AS taverncord

LABEL org.opencontainers.image.source="https://github.com/wrapped-owls/gontainer_foundryvtt"
LABEL org.opencontainers.image.description="TavernCord Discord bot for FoundryVTT"

RUN apk add --no-cache ca-certificates tzdata

COPY --from=bot-build /out/taverncord /usr/local/bin/taverncord

ENTRYPOINT ["taverncord"]

FROM builder AS foundry-build

RUN nix build .#runtimeDeps --no-link
COPY . .
RUN nix build .#runtimeRoot --print-out-paths --no-link > /tmp/root-path \
 && nix build .#runtimeDeps --print-out-paths --no-link > /tmp/deps-path \
 && nix-store -qR "$(cat /tmp/deps-path)" | sort > /tmp/deps \
 && nix-store -qR "$(cat /tmp/root-path)" | sort > /tmp/root \
 && mkdir -p /out/deps /out/app /out/root \
 && cp -a $(comm -12 /tmp/deps /tmp/root) /out/deps/ \
 && cp -a $(comm -13 /tmp/deps /tmp/root) /out/app/ \
 && cp -a "$(cat /tmp/root-path)"/. /out/root/

FROM scratch AS runtime

LABEL org.opencontainers.image.source="https://github.com/wrapped-owls/gontainer_foundryvtt"
LABEL org.opencontainers.image.description="FoundryVTT container runtime"

COPY --from=foundry-build /out/deps /nix/store
COPY --from=foundry-build /out/app /nix/store
COPY --from=foundry-build /out/root /

ENV PATH=/bin \
    LD_LIBRARY_PATH=/lib \
    FOUNDRY_DATA_PATH=/data \
    FOUNDRY_INSTALL_ROOT=/foundry \
    FOUNDRY_SOURCES_DIR=/foundry/sources \
    FOUNDRY_PATCH_MANIFEST=/etc/foundry/patches/manifest.yaml \
    SSL_CERT_FILE=/etc/ssl/certs/ca-bundle.crt \
    TZDIR=/etc/zoneinfo

VOLUME ["/data"]
EXPOSE 30000/tcp

HEALTHCHECK --start-period=3m --interval=30s --timeout=5s \
  CMD ["/bin/foundryctl", "healthcheck"]

ENTRYPOINT ["/bin/foundryctl"]
CMD ["run"]
