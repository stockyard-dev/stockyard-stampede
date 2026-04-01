# Stockyard Stampede

**Load tester.** Define a target, set concurrency and duration, fire. Real-time RPS, latency percentiles, error rate. Like k6 but a single binary with a dashboard. No external dependencies.

Part of the [Stockyard](https://stockyard.dev) suite of self-hosted developer tools.

## Quick Start

```bash
curl -sfL https://stockyard.dev/install/stampede | sh
stampede
```

Dashboard at [http://localhost:8880/ui](http://localhost:8880/ui)

## Usage

```bash
# Create a test
curl -X POST http://localhost:8880/api/tests \
  -H "Content-Type: application/json" \
  -d '{"name":"API Health","url":"https://api.example.com/health","concurrency":50,"duration_seconds":30}'

# Run it
curl -X POST http://localhost:8880/api/tests/{id}/run

# Watch live stats (poll while running)
curl http://localhost:8880/api/runs/{run_id}/live
# → {"running":true,"total":4523,"rps":"150.8","success":4520,"errors":3}

# Get final results with percentiles
curl http://localhost:8880/api/runs/{run_id}
# → {"rps":150.77,"p50_ms":12,"p95_ms":45,"p99_ms":120,...}
```

## Free vs Pro

| Feature | Free | Pro ($4.99/mo) |
|---------|------|----------------|
| Tests | 5 | Unlimited |
| Max concurrency | 10 | 500 |
| Max duration | 30s | 10 min |
| Run history | 7 days | 90 days |

## License

Apache 2.0 — see [LICENSE](LICENSE).
