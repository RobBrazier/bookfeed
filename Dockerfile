FROM cgr.dev/chainguard/bash:latest
ARG TARGETPLATFORM
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --interval=30s --timeout=2s --start-period=5s --retries=3 CMD [ "curl", "http://localhost:8080/up"]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
