FROM golang:latest

ARG IRIS_DIR=/var/lib/iris
ENV IRIS_DIR=$IRIS_DIR

RUN apt-get update -y && apt-get upgrade -y \
    && apt install build-essential git -y \
    && mkdir -p $IRIS_DIR

WORKDIR ${IRIS_DIR}
COPY . .

RUN make install

COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh

ENV SHELL /bin/bash
EXPOSE 1317 26656 26657

ENTRYPOINT ["entrypoint.sh"]
