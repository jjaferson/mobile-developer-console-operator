NAMESPACE=mobile-developer-console
CODE_COMPILE_OUTPUT = build/_output/bin/mobile-developer-console-operator
TEST_COMPILE_OUTPUT = build/_output/bin/mobile-developer-console-operator-test

QUAY_ORG=aerogear
QUAY_IMAGE=mobile-developer-console-operator
DEV_TAG ?= $(shell sh -c "git rev-parse --short HEAD")
OPENSHIFT_HOST ?= $(shell minishift ip):8443

.PHONY: setup/travis
setup/travis:
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	@echo Installing Operator SDK
	curl -Lo ${GOPATH}/bin/operator-sdk https://github.com/operator-framework/operator-sdk/releases/download/v0.7.0/operator-sdk-v0.7.0-x86_64-linux-gnu
	chmod +x ${GOPATH}/bin/operator-sdk
	@echo setup complete

.PHONY: code/compile
code/compile: code/gen
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o=$(CODE_COMPILE_OUTPUT) ./cmd/manager/main.go

.PHONY: code/run
code/run: code/gen
ifndef OPENSHIFT_HOST
	$(error OPENSHIFT_HOST is undefined)
endif
	operator-sdk up local

.PHONY: code/gen
code/gen: code/fix
	operator-sdk generate k8s
	operator-sdk generate openapi
	go generate ./...

.PHONY: code/fix
code/fix:
	gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

.PHONY: test/unit
test/unit:
	@echo Running tests:
	CGO_ENABLED=1 go test -v -race -cover ./pkg/...

.PHONY: test/compile
test/compile:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -o=$(TEST_COMPILE_OUTPUT) ./test/e2e/...

.PHONY: cluster/prepare
cluster/prepare:
	-kubectl create namespace $(NAMESPACE)
	-kubectl label namespace $(NAMESPACE) monitoring-key=middleware
	-kubectl create -n $(NAMESPACE) -f deploy/service_account.yaml
	-kubectl create -n $(NAMESPACE) -f deploy/role.yaml
	-kubectl create -n $(NAMESPACE) -f deploy/role_binding.yaml
	-kubectl create -n $(NAMESPACE) -f deploy/mobiledeveloper_role.yaml
	-kubectl create -n $(NAMESPACE) -f deploy/mobiledeveloper_rolebinding.yaml
	-kubectl apply  -n $(NAMESPACE) -f deploy/crds/mdc_v1alpha1_mobiledeveloperconsole_crd.yaml
	-kubectl apply  -n $(NAMESPACE) -f deploy/mdc_v1alpha1_mobileclient_crd.yaml

.PHONY: cluster/clean
cluster/clean:
	make uninstall
	-kubectl delete -f deploy/role.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/role_binding.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/mobiledeveloper_role.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/mobiledeveloper_rolebinding.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/service_account.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/crds/mdc_v1alpha1_mobiledeveloperconsole_crd.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/mdc_v1alpha1_mobileclient_crd.yaml
	-kubectl delete namespace $(NAMESPACE)

.PHONY: install-operator
install-operator:
	-kubectl apply -n $(NAMESPACE) -f deploy/operator.yaml
	-kubectl set env -n $(NAMESPACE) -f deploy/operator.yaml OPENSHIFT_HOST=${OPENSHIFT_HOST}

.PHONY: install-mdc
install-mdc:
	-kubectl apply -n $(NAMESPACE) -f deploy/crds/mdc_v1alpha1_mobiledeveloperconsole_cr.yaml

.PHONY: uninstall
uninstall:
	-kubectl delete -n $(NAMESPACE) MobileDeveloperConsole --all

.PHONY: image/build
image/build:
	operator-sdk build quay.io/${QUAY_ORG}/${QUAY_IMAGE}:${DEV_TAG} --image-build-args "--label quay.expires-after=2w"

.PHONY: image/push
image/push: image/build
	docker push quay.io/${QUAY_ORG}/${QUAY_IMAGE}:${DEV_TAG}

.PHONY: monitoring/install
monitoring/install:
	@echo Installing service monitor in ${NAMESPACE} :
	- kubectl label namespace ${NAMESPACE} monitoring-key=middleware
	- kubectl apply -n ${NAMESPACE} -f deploy/monitor/service_monitor.yaml
	- kubectl apply -n ${NAMESPACE} -f deploy/monitor/prometheus_rule.yaml
	- kubectl apply -n ${NAMESPACE} -f deploy/monitor/grafana_dashboard.yaml
	- kubectl apply -n ${NAMESPACE} -f deploy/monitor/mdc_service_monitor.yaml
	- kubectl apply -n ${NAMESPACE} -f deploy/monitor/mdc_prometheus_rule.yaml
	- kubectl apply -n ${NAMESPACE} -f deploy/monitor/mdc_grafana_dashboard.yaml

.PHONY: monitoring/uninstall
monitoring/uninstall:
	@echo Uninstalling monitor service from ${NAMESPACE} :
	- kubectl delete -n ${NAMESPACE} -f deploy/monitor/service_monitor.yaml
	- kubectl delete -n ${NAMESPACE} -f deploy/monitor/prometheus_rule.yaml
	- kubectl delete -n ${NAMESPACE} -f deploy/monitor/grafana_dashboard.yaml
	- kubectl delete -n ${NAMESPACE} -f deploy/monitor/mdc_service_monitor.yaml
	- kubectl delete -n ${NAMESPACE} -f deploy/monitor/mdc_prometheus_rule.yaml
	- kubectl delete -n ${NAMESPACE} -f deploy/monitor/mdc_grafana_dashboard.yaml
