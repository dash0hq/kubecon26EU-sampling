# Head sampling

Head sampling discards traces by a fixed probability, evaluated deterministically from the trace ID.
Every Collector instance reaches the same decision for the same trace without coordination, so it scales horizontally without any special routing.

## Trade-off

The sampling decision is made before the outcome of the request is known.
At low sampling rates, errors, latency spikes, and rarely exercised code paths are likely to be missed.
Use head sampling when tail sampling is too complex for the environment, or as a coarse volume reduction stage before tail sampling.

## Architecture

![Head sampling architecture](../assets/diagrams/head-sampling.svg)

Because the sampling decision is deterministic from the trace ID, all Collector replicas reach the same conclusion — no load-balancing tier is needed.

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
kubectl apply -f collector.yaml
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

| File             | Description                                                          |
|------------------|----------------------------------------------------------------------|
| `jaeger.yaml`    | Jaeger all-in-one — in-memory backend, OTLP receiver, UI            |
| `collector.yaml` | ConfigMap, Deployment, and Service — runs `probabilistic_sampler`   |
| `apps.yaml`      | `node-frontend`, `go-backend`, and `load-generator` Deployments     |
