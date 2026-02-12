FROM cgr.dev/chainguard/busybox:latest@sha256:76a4b21bc9841854086c39ba4b14b06cc8a9f826838ff08d4a08b3a40900fcf4
ARG TARGETPLATFORM
USER root
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --start-period=5s CMD [ "pgrep", "bookfeed" ]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
