FROM gcr.io/distroless/static-debian13:nonroot

ARG TARGETPLATFORM
COPY --chmod=0555 ${TARGETPLATFORM}/catl /catl

ENV XDG_CACHE_HOME=/tmp/.cache \
    CATL_CACHE_DIR=/tmp/.cache/catl

USER 65532:65532
ENTRYPOINT ["/catl"]