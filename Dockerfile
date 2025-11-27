FROM cgr.dev/chainguard/busybox:latest@sha256:ecc152fe3dece44e60d1aa0fbbefb624902b4af0e2ed8c2c84dfbce653ff064f
ARG TARGETPLATFORM
USER root
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --start-period=5s CMD [ "pgrep", "bookfeed" ]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
