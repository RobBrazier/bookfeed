FROM cgr.dev/chainguard/busybox:latest@sha256:e4d4487d5a4d1b381e0ddef556cc42be59ed435c2b0891c12c44bb550eb3a5f4
ARG TARGETPLATFORM
USER root
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --start-period=5s CMD [ "pgrep", "bookfeed" ]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
