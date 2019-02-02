FROM golang:alpine AS builder

LABEL maintainer="Javier Lopez Lopez<sjavierlopez@gmail.com>"

ENV GOUSER_HOME    /home/gouser/
ARG AWS_ACCESS_KEY_ID=xxx
ARG AWS_SECRET_ACCESS_KEY=xxx
ARG user=gouser
ARG group=gouser
ARG uid=1000
ARG gid=1000

RUN apk update \
    && apk --update --no-cache add bash  git \
    && mkdir -p ${GOPATH}/src ${GOPATH}/bin


WORKDIR ${GOPATH}/src/app-platform
COPY . .

RUN cd ${GOPATH}/src/app-platform/cmd \ 
    && go get -d -v ./... \ 
    && go build -o /usr/bin/orquesta ${GOPATH}/src/app-platform/cmd/orquesta/main.go \
    && go build -o /usr/bin/deploy ${GOPATH}/src/app-platform/cmd/deploy/main.go \
    && mkdir -p $(dirname $GOUSER_HOME) \
    && mkdir -p /home/gouser/.aws/

RUN addgroup -g ${gid} ${group}
RUN adduser -h ${GOUSER_HOME} -S ${user} -u ${uid} -G ${group} -s /bin/bash ${user} 

RUN touch /home/gouser/.aws/credentials \
    && echo "[default]" >> /home/gouser/.aws/credentials \
    && echo "aws_access_key_id = ${AWS_ACCESS_KEY_ID}" >> /home/gouser/.aws/credentials \
    && echo "aws_secret_access_key = ${AWS_SECRET_ACCESS_KEY}" >> /home/gouser/.aws/credentials

RUN chown -R gouser:gouser  /home/gouser/

USER gouser