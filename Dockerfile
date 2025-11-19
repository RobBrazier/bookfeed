FROM cgr.dev/chainguard/bash:latest
ARG TARGETPLATFORM
ENTRYPOINT ["/usr/bin/bookfeed"]
COPY $TARGETPLATFORM/bookfeed /usr/bin/
