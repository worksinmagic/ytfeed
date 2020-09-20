# The compiling is done by goreleaser
FROM ubuntu:focal
RUN apt-get update
RUN apt-get -y install youtube-dl
COPY ytfeed /
ENTRYPOINT ["/ytfeed"]
