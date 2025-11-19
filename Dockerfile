FROM scratch
ARG TARGETPLATFORM
ENTRYPOINT ["/usr/bin/bookfeed"]
COPY $TARGETPLATFORM/bookfeed /usr/bin/
