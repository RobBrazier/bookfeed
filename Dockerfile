FROM cgr.dev/chainguard/busybox:latest@sha256:bfdc41993cd6bf870b453be97ae8e91fd238b652d2959e1827e221b2b3076792
ARG TARGETPLATFORM
USER root
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --start-period=5s CMD [ "pgrep", "bookfeed" ]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
