FROM cgr.dev/chainguard/busybox:latest@sha256:48561a67b5d7120d6fdaeaa8b6f7d9d251a5665f90b5e05b1dc1be4c17850db4
ARG TARGETPLATFORM
USER root
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --start-period=5s CMD [ "pgrep", "bookfeed" ]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
