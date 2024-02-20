_default:
    just --list

alias start-nes := start-network-extended-scheduler

[private]
alias start-default-scheduler := start-default

arch := replace(replace(arch(), "aarch64", "arm64v8"), "x86_64", "amd64")
kwokNodes := "200"
scheduler := "network-extended-scheduler"

# --- Start Environment ---

# Start a local jupyter-lab instance for data analysis
lab:
	jupyter-lab

# Start KWOK-managed cluster with the default scheduler
start-default: start-kwok-cluster (setup-kwok-nodes kwokNodes)

# Start KWOK-managed cluster with NES
start-network-extended-scheduler:
    just start-registry

    cd scheduler-plugins && \
    	docker build -f ./build/scheduler/Dockerfile --build-arg ARCH="{{ arch }}" --build-arg RELEASE_VERSION="v$(date +%Y%m%d)-v0.27.0" -t k3d-default-registry.localhost:9090/scheduler-plugins/kube-scheduler:latest-{{ arch }} . && \
    	docker push k3d-default-registry.localhost:9090/scheduler-plugins/kube-scheduler:latest-{{ arch }}

    just start-kwok-cluster "k3d-default-registry.localhost:9090/scheduler-plugins/kube-scheduler:latest-{{ arch }}" "scheduler-plugins/manifests/networkbandwidth/scheduler-config.yaml"

    just setup-kwok-nodes {{ kwokNodes }}

# Start KWOK-managed cluster with an explicitly defined default scheduler
[private]
start-default-kubernetes-scheduler: (start-kwok-cluster "registry.k8s.io/kube-scheduler:v1.27.1" "scheduler-plugins/manifests/default/scheduler-config.yaml") (setup-kwok-nodes kwokNodes)

# Stops the local test cluster
[no-exit-message]
stop:
    -just stop-registry
    -just stop-kwok-cluster

# --- Test Scenarios ---

# Generate scheduler files
[private]
generate-scheduler-files:
    node ./manifests/scenarios.js

# Build measurements collector
[private]
build-measurements-collector:
    go build -o collector measurements.go

# Remove built binary
[private]
remove-measurements-collector:
    rm collector

# Scenario 1: Fixed and strict network bandwidth requirements
scenario-1: generate-scheduler-files build-measurements-collector
    kubectl apply -f ./manifests/{{ scheduler }}/1500-100Mi-100Mi.yaml
    ./collector -interval=1s -output=./measurements/scenario-1/{{ scheduler }}.csv

# Scenario 2: Burstable network bandwidth requirements and node overcommitment
scenario-2: generate-scheduler-files build-measurements-collector
    kubectl apply -f ./manifests/{{ scheduler }}/1500-50Mi-100Mi.yaml
    ./collector -interval=1s -output=./measurements/scenario-2/{{ scheduler }}.csv

# Scenario 3: Workload distribution with mixed workloads
scenario-3: generate-scheduler-files build-measurements-collector
    kubectl apply -f ./manifests/{{ scheduler }}/1500-100Mi-100Mi-cpu.yaml
    ./collector -interval=1s -output=./measurements/scenario-3/{{ scheduler }}.csv

# Scenario 4: Workload distribution with mixed, burstable workloads
scenario-4: generate-scheduler-files build-measurements-collector
    kubectl apply -f ./manifests/{{ scheduler }}/1500-50Mi-100Mi-cpu.yaml
    ./collector -interval=1s -output=./measurements/scenario-4/{{ scheduler }}.csv

# Perform a scenario test with a specific scheduler
perform-scenario scenario="1": stop
    just kwokNodes={{kwokNodes}} arch={{arch}} start-{{ scheduler }}
    just --set scheduler {{scheduler}} scenario-{{ scenario }}

# --- K3d ---

# Setup local oci registry
[private]
start-registry:
    k3d registry create default-registry.localhost --port 9090

# Delete local oci registry
[private]
[no-exit-message]
stop-registry:
    k3d registry delete default-registry.localhost

# --- KWOK ---

# Start a KWOK cluster
[private]
start-kwok-cluster kube-scheduler-image="registry.k8s.io/kube-scheduler:v1.27.1" kube-scheduler-config="":
    kwokctl create cluster \
    	--name=kwok \
    	--kube-scheduler-image={{ kube-scheduler-image }} \
    	--kube-scheduler-config={{ kube-scheduler-config }} \
    	--prometheus-image=docker.io/prom/prometheus:v2.43.0 \
    	--prometheus-port=9091

# Delete a KWOK cluster
[private]
[no-exit-message]
stop-kwok-cluster:
    kwokctl delete cluster --name=kwok

# Setup KWOK nodes
[private]
setup-kwok-nodes nodes="10":
    node ./manifests/kwok/node.js {{ nodes }} | kubectl apply -f -
