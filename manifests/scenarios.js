#!/usr/bin/env node

const fs = require("fs");

const maxPods = [1500];
const bandwidth = [
	[100, 100],
	[50, 100],
];

const schedulerNames = [
	"default-scheduler",
	"default-kubernetes-scheduler",
	"network-extended-scheduler",
	// "kube-scheduler-rs",
];

const deployment = (
	max,
	schedulerName,
	bandwidth,
	bandwidthLimit = bandwidth,
) => `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stress-test
  namespace: test-${schedulerName}-${max}-${bandwidth}
spec:
  replicas: ${max}
  selector:
    matchLabels:
      app: pause
  template:
    metadata:
      annotations:
        kubernetes.io/ingress-bandwidth: ${bandwidthLimit}M
        kubernetes.io/egress-bandwidth: ${bandwidthLimit}M
        kubernetes.io/ingress-request: ${bandwidth}M
        kubernetes.io/egress-request: ${bandwidth}M
      labels:
        app: pause
    spec:
      schedulerName: ${schedulerName}
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: type
                    operator: In
                    values:
                      - kwok
      tolerations:
        - key: "kwok.x-k8s.io/node"
          operator: "Exists"
          effect: "NoSchedule"
      containers:
        - name: pause-container
          image: gcr.io/google_containers/pause:3.2
          resources:
            requests:
              cpu: "0.1"
              memory: "10Mi"
            limits:
              cpu: "0.1"
              memory: "10Mi"
`;

const cpuDeployment = (schedulerName, bw, maxBw) => {
	const burstable = bw != maxBw;

	return `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stress-test
  namespace: cpu-workload-${schedulerName}-${
		burstable ? "burstable" : "guaranteed"
	}
spec:
  replicas: 1600
  selector:
    matchLabels:
      app: cpu-memory
  template:
    metadata:
      annotations:
        kubernetes.io/ingress-bandwidth: "0"
        kubernetes.io/egress-bandwidth: "0"
        kubernetes.io/ingress-request: "0"
        kubernetes.io/egress-request: "0"
      labels:
        app: cpu-memory
    spec:
      schedulerName: ${schedulerName}
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: type
                    operator: In
                    values:
                      - kwok
      tolerations:
        - key: "kwok.x-k8s.io/node"
          operator: "Exists"
          effect: "NoSchedule"
      containers:
        - name: pause-container
          image: gcr.io/google_containers/pause:3.2
          resources:
            requests:
              cpu: "${burstable ? 1 : 2}"
              memory: "${burstable ? "8Gi" : "16Gi"}"
            limits:
              cpu: "2"
              memory: "16Gi"`;
};

const ns = (
	max,
	schedulerName,
	bandwidth,
	bandwidthLimit = bandwidth,
) => `apiVersion: v1
kind: Namespace
metadata:
  labels:
    kubernetes.io/metadata.name: test-${schedulerName}-${max}-${bandwidth}-${bandwidthLimit}
  name: test-${schedulerName}-${max}-${bandwidth}-${bandwidthLimit}`;

for (const schedulerName of schedulerNames) {
	fs.mkdirSync(`./manifests/${schedulerName}`, { recursive: true });

	for (const otherWorkload of [false, true]) {
		for (const max of maxPods) {
			for (const [bw, maxBw] of bandwidth) {
				let fileName = `./manifests/${schedulerName}/${max}-${bw}Mi-${maxBw}Mi`;
				if (otherWorkload) {
					fileName += "-cpu";
				}
				fileName += ".yaml";

				let definition = `${ns(max, schedulerName, bw, maxBw)}
${deployment(max, schedulerName, bw, maxBw)}`;

				if (otherWorkload) {
					definition += cpuDeployment(schedulerName, bw, maxBw);
				}

				fs.writeFileSync(fileName, definition);
			}
		}
	}
}
