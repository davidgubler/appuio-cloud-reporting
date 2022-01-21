.PHONY: build
build:
	helm template graphs --namespace=appuio-reporting grafana/grafana \
		-f grafana-helm-values.yaml \
		> grafana.yaml

.PHONY: update
update:
	helm repo update
	helm search repo grafana
