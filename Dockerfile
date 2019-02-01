FROM golang:alpine AS builder

LABEL maintainer="Javier Lopez Lopez<sjavierlopez@gmail.com>"


RUN apk update \
    && apk add --no-cache git \
    && mkdir -p ${GOPATH}/src ${GOPATH}/bin


WORKDIR ${GOPATH}/src/app-platform
COPY . .

RUN cd ${GOPATH}/src/app-platform/cmd \ 
    && go get -d -v ./... \ 
    && go build -o ${GOPATH}/bin/orquesta ${GOPATH}/src/app-platform/cmd/orquesta/main.go \
    && go build -o ${GOPATH}/bin/deploy ${GOPATH}/src/app-platform/cmd/deploy/main.go

RUN adduser -D -g '' gouser    

FROM ubuntu

ENV GOUSER_HOME          /home/gouser/

ARG user=gouser
ARG group=gouser
ARG uid=1000
ARG gid=1000

RUN mkdir -p $(dirname $GOUSER_HOME) \
    && groupadd -g ${gid} ${group} \
    && useradd -d "$GOUSER_HOME" -u ${uid} -g ${gid} -m -s /bin/bash ${user}

COPY --from=builder /go/bin/deploy  /usr/bin/deploy
COPY --from=builder /go/bin/orquesta  /usr/bin/orquesta
COPY config/aws/config  /home/gouser/.aws/config
COPY config/aws/credentials  /home/gouser/.aws/credentials

RUN chown -R gouser:gouser  /home/gouser/


USER gouser