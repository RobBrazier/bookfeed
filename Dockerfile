FROM cgr.dev/chainguard/busybox:latest@sha256:09c1dcec5d6fa044651a572081f3a33a0d84654168c8e3ea2fc5a07913e41d74
ARG TARGETPLATFORM
USER root
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --start-period=5s CMD [ "pgrep", "bookfeed" ]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
