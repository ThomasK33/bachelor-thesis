# Bachelors Thesis

This is the accompanying source code to the thesis "Extending the Kubernetes scheduler for network resource awareness" by Thomas Kosiewski.

## Contributions and changes

- Changes performed in scheduler-plugins project to integrate code are contained in `scheduler-plugins.patch` git patch file.
- The scheduler extension is located at: `scheduler-plugins/pkg/networkaware/networkbandwidth/plugin.go`
- Scheduler configuration file for NES and the default scheduler: `scheduler-plugins/manifests/networkbandwidth/scheduler-config.yaml` and `scheduler-plugins/manifests/default/scheduler-config.yaml`
- Metrics collection utility: `measurements.go`.
  - `measurements` contains the metrics collected for each scheduling scenario.
- Jupyter Notebook for data analysis and visualization: `figures.ipynb`.
- Justfile - Collection of scripts and tasks: `justfile`
- `manifests` contains the files for each scheduling scenario:
  - KWOK node manifest generation script at `manifests/kwok/node.js`
  - Scenario manifest generation script located at: `manifests/scenario.js`

## Prerequisites for running this project

The following applications are needed to run this project

- [just](https://github.com/casey/just)
- [node.js](https://nodejs.org/en)
- [docker & docker compose](https://www.docker.com)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [k3d](https://k3d.io/v5.5.1/)
- [kwokctl](https://kwok.sigs.k8s.io)
- [Go](https://go.dev)
- make

## Running a cluster

Run `just` for a list of available recipes.

### Starting local cluster

#### With default scheduler

Run `just start-default` to start a KWOK-managed cluster with the default Kubernetes scheduler deployed.

#### With Network Extended Scheduler

Run `just start-network-extended-scheduler` to start a KWOK-managed cluster with the NES scheduler deployed.

Specifying the CPU architecture

- In order to build and run NES on a arm64 architecture, execute `just start-network-extended-scheduler 200 arm64v8`.
- In order to build and run NES on a amd64 architecture, execute `just start-network-extended-scheduler 200 amd64`.

## Running scenarios

To run a given scenario, perform one of the following recipes:

- `just scenario-1`
- `just scenario-2`
- `just scenario-3`
- `just scenario-4`

After each scenario, one has to specify which scheduler is currently deployed in the cluster.

For example, to be able to run scenario-1 in a cluster with the network extended scheduler, one will have to run: `just scenario-1 network-extended-scheduler`.
Whereas in a cluster with the default scheduler, one has to run `just scenario-1 default-scheduler`

## End-to-End instructions

### Default scheduler

In order to perform the full scenario test suite, run the following commands:

```bash
just start-default
just scenario-1 "default-scheduler" # Replace scenario-1 with any scenario listed above
just stop
```

### Network Extended Scheduler

In order to perform the full scenario test suite, run the following commands:

```bash
just start-network-extended-scheduler
just scenario-1 # Replace scenario-1 with any scenario listed above
just stop
```
