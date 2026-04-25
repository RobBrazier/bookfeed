FROM cgr.dev/chainguard/busybox:latest@sha256:8b6bf2c4e57354e9a3d184bafa893b40e93ceaa1b48560c31c2c4d8a318f64e1
ARG TARGETPLATFORM
USER root
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --start-period=5s CMD [ "pgrep", "bookfeed" ]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
