ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest@sha256:f173c44fab35484fa0e940e42929efe2a2f506feda431ba72c5f0d79639d7f55

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/databricks-exporter /bin/databricks-exporter

EXPOSE      9975
USER        nobody
ENTRYPOINT  [ "/bin/databricks-exporter" ]
