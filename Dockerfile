FROM gcr.io/distroless/static-debian13:nonroot
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/catl /catl
USER nonroot:nonroot
ENTRYPOINT ["/catl"]