FROM cgr.dev/chainguard/busybox:latest@sha256:8268eddcce2d4c99c53ab8eaee65e3bddd4933b3e400ba0aa25509ee7573d8b8
ARG TARGETPLATFORM
USER root
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --start-period=5s CMD [ "pgrep", "bookfeed" ]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
