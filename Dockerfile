ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest@sha256:4eb7e9b67c0839f512270bae2d84fa75ad2c09075d11468cc1c1cef19b573ebc

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/databricks-exporter /bin/databricks-exporter

EXPOSE      9975
USER        nobody
ENTRYPOINT  [ "/bin/databricks-exporter" ]
