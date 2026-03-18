# OpenTelemetry sampling — KubeCon EU 2026

Companion material for the [KubeCon EU 2026 talk on OpenTelemetry sampling strategies](https://events.linuxfoundation.org/kubecon-cloudnativecon-europe/program/schedule/).
Each directory contains a self-contained Kubernetes setup for a different approach.

## Strategies

| Strategy      | Directory                          | When to use                                                  |
|---------------|------------------------------------|--------------------------------------------------------------|
| Head sampling | [`head-sampling/`](head-sampling/) | Simple, stateless volume reduction by a fixed percentage     |
| Tail sampling | [`tail-sampling/`](tail-sampling/) | Keep errors and slow traces; drop only healthy, fast traces  |

## Head sampling

The sampling decision is made upfront, before the trace completes, based on the trace ID alone.
Every Collector replica independently reaches the same decision — no coordination needed.

![Head sampling architecture](assets/diagrams/head-sampling.svg)

See [head-sampling/README.md](head-sampling/README.md).

## Tail sampling

The sampling decision is made after the full trace is collected, so errors and slow requests can always be retained regardless of sampling rate.
Requires a two-tier Collector setup.

![Tail sampling architecture](assets/diagrams/tail-sampling.svg)

See [tail-sampling/README.md](tail-sampling/README.md).

## Prerequisites

- Kubernetes cluster with `kubectl` configured

Each setup is self-contained and includes [Jaeger](https://www.jaegertracing.io) as the trace backend.
Jaeger uses in-memory storage — suitable for demos, not for production.

Depending on your cluster's setup, you may need to use port-forwarding or an ingress controller to expose Jaeger.

## Deploy both scenarios side by side

The two scenarios can run simultaneously in separate namespaces.
The commands below substitute the namespace in every manifest on the fly, so no files need to be edited.

### 1. Build the application images

```bash
docker build -t node-frontend:latest apps/node-frontend
docker build -t go-backend:latest    apps/go-backend
```

Load them into your local cluster if you are using [kind](https://kind.sigs.k8s.io):

```bash
kind load docker-image node-frontend:latest
kind load docker-image go-backend:latest
```

### 2. Deploy head sampling

```bash
kubectl create namespace head-sampling

for f in head-sampling/*.yaml; do
  sed 's/namespace: otel/namespace: head-sampling/g
       s/namespace: default/namespace: head-sampling/g
       s/\.otel\.svc\.cluster\.local/.head-sampling.svc.cluster.local/g' "$f" \
  | kubectl apply -f -
done
```

### 3. Deploy tail sampling

```bash
kubectl create namespace tail-sampling

for f in tail-sampling/*.yaml; do
  sed 's/namespace: otel/namespace: tail-sampling/g
       s/namespace: default/namespace: tail-sampling/g
       s/\.otel\.svc\.cluster\.local/.tail-sampling.svc.cluster.local/g' "$f" \
  | kubectl apply -f -
done
```

### 4. Open the Jaeger UIs

```bash
# Head sampling — http://localhost:16686
kubectl port-forward -n head-sampling svc/jaeger 16686:16686 &

# Tail sampling — http://localhost:16687
kubectl port-forward -n tail-sampling svc/jaeger 16687:16686 &
```

The [load generator](apps/) sends one request per second to each scenario automatically.
After a few minutes, both Jaeger instances will show sampled traces from the `node-frontend` and `go-backend` services.

### 5. Tear down

```bash
kubectl delete namespace head-sampling
kubectl delete namespace tail-sampling
```
