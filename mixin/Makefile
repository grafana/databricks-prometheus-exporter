.PHONY: all clean build fmt lint install update

JSONNET_FMT := jsonnetfmt -n 2 --max-blank-lines 2 --string-style s --comment-style s

all: build dashboards_out prometheus_alerts.yaml

vendor: jsonnetfile.json
	jb install

build: vendor

fmt:
	find . -name 'vendor' -prune -o -name '*.libsonnet' -print -o -name '*.jsonnet' -print | \
		xargs -n 1 -- $(JSONNET_FMT) -i

lint: build
	find . -name 'vendor' -prune -o -name '*.libsonnet' -print -o -name '*.jsonnet' -print | \
		while read f; do \
			$(JSONNET_FMT) "$$f" | diff -u "$$f" -; \
		done
	mixtool lint mixin.libsonnet

install:
	jb install

update:
	jb update

dashboards_out: mixin.libsonnet config.libsonnet
	@mkdir -p dashboards_out
	@echo "Generating dashboards..."
	@if command -v mixtool >/dev/null 2>&1; then \
		mixtool generate dashboards mixin.libsonnet -d dashboards_out; \
	else \
		jsonnet -J vendor -m dashboards_out mixin.libsonnet; \
	fi

prometheus_alerts.yaml: mixin.libsonnet alerts.jsonnet
	@echo "Generating Prometheus alerts..."
	@if command -v mixtool >/dev/null 2>&1; then \
		mixtool generate alerts mixin.libsonnet -a prometheus_alerts.yaml; \
	else \
		jsonnet -J vendor -S -y alerts.jsonnet > prometheus_alerts.yaml; \
	fi

clean:
	rm -rf dashboards_out prometheus_alerts.yaml

