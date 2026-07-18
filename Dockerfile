FROM cgr.dev/chainguard/busybox:latest
ARG TARGETPLATFORM
ENTRYPOINT ["/app/bookfeed"]
COPY assets/build /app/static
COPY $TARGETPLATFORM/bookfeed /app
