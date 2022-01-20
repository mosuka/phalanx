FROM golang:1.17.5-bullseye

ARG TAG

ENV GOPATH /go

COPY . ${GOPATH}/src/github.com/mosuka/phalanx

RUN apt-get update \
    && apt-get install -y \
       build-essential \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN cd ${GOPATH}/src/github.com/mosuka/phalanx \
    && make TAG=${TAG} build


FROM debian:bullseye-slim

RUN apt-get update \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r phalanx && useradd -r -g phalanx phalanx
USER phalanx

COPY --from=0 --chown=phalanx:phalanx /go/src/github.com/mosuka/phalanx/bin/* /usr/bin/

EXPOSE 2000 5000 8000

ENTRYPOINT [ "/usr/bin/phalanx" ]
