FROM cgr.dev/chainguard/bash:latest@sha256:b8210a1e571844b4839347864086a50a12bdf6b160636d9d5b4d349d06c9b670
ARG TARGETPLATFORM
ENTRYPOINT ["/usr/bin/bookfeed"]
HEALTHCHECK --interval=30s --timeout=2s --start-period=5s --retries=3 CMD [ "curl", "http://localhost:8080/up"]
EXPOSE 8080
COPY $TARGETPLATFORM/bookfeed /usr/bin/
