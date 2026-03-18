# Tail sampling

Tail sampling requires a two-tier architecture.
Single-tier setups with `replicaCount > 1` are broken: spans from the same trace can land on different pods, so sampling decisions are made on incomplete data.

## Architecture

![Tail sampling architecture](../assets/diagrams/tail-sampling.svg)

The agent routes all spans of a trace to the same gateway pod via the `loadbalancingexporter`.
The gateway then buffers the full trace and applies sampling policies.

A headless Service (`clusterIP: None`) is required for the gateway so that DNS returns individual pod IPs — the load-balancing exporter needs these to maintain a consistent hash ring.

## Build the application images

```bash
docker build -t node-frontend:latest ../apps/node-frontend
docker build -t go-backend:latest ../apps/go-backend
```

Load them into your cluster (example for [kind](https://kind.sigs.k8s.io)):

```bash
kind load docker-image node-frontend:latest
kind load docker-image go-backend:latest
```

## Deploy

```bash
kubectl create namespace otel

kubectl apply -f jaeger.yaml
kubectl apply -f headless-service.yaml
kubectl apply -f collector-gateway.yaml
kubectl apply -f collector-agent.yaml
kubectl apply -f apps.yaml
```

Access the Jaeger UI:

```bash
kubectl port-forward svc/jaeger 16686:16686 -n otel
```

Then open `http://localhost:16686`.
The [load-generator](apps.yaml) sends one request per second automatically.

## Applications

The two demo applications produce traces with a mix of outcomes to exercise the sampling policies:

| App | Language | Role |
|-----|----------|------|
| `node-frontend` | Node.js | Receives HTTP requests, calls `go-backend`, returns the result |
| `go-backend` | Go | Rolls a dice (1–6); value 1 returns a 500 error, value 6 takes 2.5 s |

Instrumentation is zero-code for Node.js (`NODE_OPTIONS=--require ...`) and uses the Go OTel SDK directly.
Both services use `AlwaysOn` sampling — all sampling decisions are made in the Collector.

## Files

| File                      | Description                                                               |
|---------------------------|---------------------------------------------------------------------------|
| `jaeger.yaml`             | Jaeger all-in-one — in-memory backend, OTLP receiver, UI                 |
| `headless-service.yaml`   | Headless Service for gateway DNS discovery by the load-balancing exporter |
| `collector-gateway.yaml`  | ConfigMap, Deployment, and Service — runs `tailsamplingprocessor`         |
| `collector-agent.yaml`    | ConfigMap, DaemonSet, and Service — routes traces via `loadbalancingexporter` |
| `apps.yaml`               | `node-frontend`, `go-backend`, and `load-generator` Deployments           |
