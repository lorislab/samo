FROM debian:10.7-slim AS builder
ARG VERSION=1.0.0

ENV FILENAME=samo_${VERSION}_Linux_x86_64.tar.gz

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl ca-certificates

RUN curl https://github.com/lorislab/samo/releases/download/${VERSION}/${FILENAME} -O -J -L && \
    tar xfz $FILENAME samo && \
    chmod +x samo

FROM debian:10.7-slim

LABEL org.opencontainers.image.source https://github.com/lorislab/samo

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder samo /opt/samo

ENTRYPOINT ["/opt/samo"]