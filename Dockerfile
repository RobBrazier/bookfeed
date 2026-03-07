FROM cgr.dev/chainguard/busybox:latest@sha256:9a0c4b6370eb2bb961edea36bf6551b05678eb34b2cb80d6b5f5a8442f2e3df3
ARG TARGETPLATFORM
USER root
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --start-period=5s CMD [ "pgrep", "bookfeed" ]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
