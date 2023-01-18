FROM alpine/helm:3.10.2 as helm

FROM debian:10.11-slim

COPY --from=helm /usr/bin/helm /usr/bin/helm

RUN apt-get update \
    && apt-get install -y apt-transport-https ca-certificates curl gnupg2 software-properties-common curl ca-certificates \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add - \
    && add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable" \
    && apt-get update \
    && apt-get install -y docker-ce

COPY samo /usr/bin/samo
