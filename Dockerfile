# The compiling is done by goreleaser
FROM ubuntu:focal
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update
RUN apt-get -y install youtube-dl ffmpeg
COPY ytfeed /
ENTRYPOINT ["/ytfeed"]
