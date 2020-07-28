ARG BASE_IMAGE_BUILD=golang:alpine
ARG BASE_IMAGE_RELEASE=alpine:latest

FROM ${BASE_IMAGE_BUILD} AS build-env
COPY . /src/
RUN cd /src \
    && apk add --no-cache git \
    && go get github.com/gin-gonic/gin \
    && go build -o bankdataservice

FROM ${BASE_IMAGE_RELEASE}
COPY --from=build-env /src/bankdataservice /usr/local/bin/
RUN mkdir -p /data
ENV DATADIRECTORY=/data \
    GIN_MODE=release
CMD ["/usr/local/bin/bankdataservice"]
