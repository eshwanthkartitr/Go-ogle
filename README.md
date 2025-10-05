# Go-ogle: A Distributed Search Engine MVP

This project is a production-minded MVP of a distributed search engine tailored for preparing for Google Software Engineer interviews. It now showcases Kafka-backed ingestion, BM25 + semantic reranking, Prometheus-first observability, and container/Kubernetes deployment assets. The codebase is organized into independently deployable services written in Go to demonstrate distributed-system design skills.

## Features

- Concurrent crawler with adaptive politeness and duplicate URL detection
- Kafka-backed ingestion pipeline decoupling crawler, indexer, and API services
- Inverted index with document store, BM25 scoring, and ANN-powered semantic reranking
- HTTP search API with Prometheus metrics and query latency histograms
- Docker and Kubernetes manifests plus unit tests to validate indexing, ranking, and semantic behaviors

## Project Layout

```
Distributed-Search-Engine/
  cmd/
    orchestrator/        # Kafka producer that crawls seed pages and publishes documents
    indexer/             # Consumes Kafka, updates indices, and writes snapshots
    searchapi/           # Exposes /search endpoint and consumes Kafka for realtime updates
  internal/
    api/                 # HTTP server wiring
    crawler/             # URL frontier management and fetching logic
    docs/                # Document model shared across services
    index/               # Inverted index, document store, and ranking helpers
    pipeline/            # Kafka integrations and index updater utilities
    search/              # Query execution with BM25 + semantic reranker
    semantic/            # Hashing embeddings and ANN index
    telemetry/           # Logging and Prometheus helpers
  deploy/
    k8s/                 # Kubernetes manifests for Kafka + services + Prometheus
  testdata/
    pages/               # Sample HTML corpus used by orchestrator tests
  ops/
    prometheus/          # Prometheus scrape configuration for docker-compose
```

## Quickstart

### Local development with Docker Compose

```bash
docker compose up --build
```

This starts Kafka, runs the crawler once to seed the topic, keeps the indexer and search API running, and launches Prometheus at `http://localhost:9090`. Query the API at `http://localhost:8080/search?q=vector+search` and inspect metrics at `http://localhost:9102/metrics`.

### Running services manually

1. **Start Kafka** – either via Docker (`docker compose up kafka`) or a local installation.
2. **Seed documents**

  ```bash
  KAFKA_BROKERS=localhost:9092 go run ./cmd/orchestrator
  ```

3. **Start the indexer**

  ```bash
  KAFKA_BROKERS=localhost:9092 go run ./cmd/indexer
  ```

4. **Start the search API**

  ```bash
  KAFKA_BROKERS=localhost:9092 go run ./cmd/searchapi
  ```

5. **Run tests**

  ```bash
  go test ./...
  ```

## Next Steps

- Integrate transformer-based rerankers (e.g., ColBERT) and deploy a vector database such as Milvus
- Add authn/z with API keys and rate-limiting at the edge tier
- Expand Kubernetes manifests with HPA rules and service meshes for resilience tests
- Wire Grafana dashboards and alert rules on top of the exported Prometheus metrics
# Go-ogle
