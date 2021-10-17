FROM debian:10.7-slim AS builder

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl ca-certificates

FROM debian:10.7-slim

LABEL org.opencontainers.image.source https://github.com/lorislab/samo

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY samo /opt/samo

ENTRYPOINT ["/opt/samo"]