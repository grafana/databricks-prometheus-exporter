ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/databricks-exporter /bin/databricks-exporter

EXPOSE      9975
USER        nobody
ENTRYPOINT  [ "/bin/databricks-exporter" ]
