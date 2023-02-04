ARG GO_VERSION="1.15.14"
ARG ALPINE_VERSION="3.15.0"

FROM alpine:$ALPINE_VERSION

COPY _build/eck-escluster-eru-limit-exporter /usr/bin/eck-escluster-eru-limit-exporter 

EXPOSE 8889

ENTRYPOINT ["/usr/bin/eck-escluster-eru-limit-exporter"]