FROM cgr.dev/chainguard/busybox:latest@sha256:6202d571cc9c86302d4bd1eef6d2235d2b81b7dfc509cf7eee621f4d73175807
ARG TARGETPLATFORM
USER root
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --start-period=5s CMD [ "pgrep", "bookfeed" ]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
