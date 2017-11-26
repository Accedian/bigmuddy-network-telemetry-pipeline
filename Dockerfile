FROM ubuntu:xenial

COPY bin/pipeline /go/bin/pipeline

ENTRYPOINT ["/go/bin/pipeline"]
