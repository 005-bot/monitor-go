<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![Apache License][license-shield]][license-url]
[![Go][go-shield]][go-url]

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/005-bot/monitor-go">
    <img src="https://raw.githubusercontent.com/golang-samples/gopher-vector/master/gopher.png" alt="Logo" width="120" height="120">
  </a>

  <h3 align="center">monitor-go</h3>

  <p align="center">
    Production-grade outage schedule monitor with Redis-backed pipeline
    <br />
    <a href="https://github.com/005-bot/monitor-go"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/005-bot/monitor-go/issues">Report Bug</a>
    ·
    <a href="https://github.com/005-bot/monitor-go/issues">Request Feature</a>
  </p>
</div>

<!-- TABLE OF CONTENTS -->
- [About The Project](#about-the-project)
- [Features](#features)
- [Built With](#built-with)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Quick Start](#quick-start)
- [Usage](#usage)
  - [Build from Source](#build-from-source)
  - [HTTP API](#http-api)
  - [Make Targets](#make-targets)
- [Configuration](#configuration)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)
- [Acknowledgments](#acknowledgments)


<!-- ABOUT THE PROJECT -->
## About The Project

`monitor-go` scrapes the Красноярск city outage schedule from [Gorod.htm](http://93.92.65.26/aspx/Gorod.htm), parses the data, detects changes, and publishes new outage records via Redis Pub/Sub. It is a modular, testable Go service that replaces a legacy Python service.

The service runs a periodic pipeline (scrape → parse → diff → publish) and also exposes an HTTP API for health checks, manual triggers, and integration with external systems.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- FEATURES -->
## Features

* **Scheduler** — periodic ticker-based orchestrator that runs the full pipeline on a configurable interval
* **Scraper** — HTTP client with ETag-based change detection, charset-aware HTML parsing (Windows-1251 via goquery)
* **Parser** — extracts organization info, Russian addresses, outage details, and Russian date strings; includes an embedded SQLite street name database for normalization
* **Storage** — Redis-backed diff engine: computes MD5 hashes of parsed records, commits only new/changed entries with TTL
* **Publisher** — serializes outages as JSON and publishes to a Redis channel (`{prefix}:outages`)
* **HTTP API** — health probes, Swagger UI, Prometheus metrics, and endpoints to trigger and inspect each pipeline stage
* **Metrics** — Prometheus counters and histograms for every operation across all modules
* **Docker** — minimal container image built via GoReleaser

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- BUILT WITH -->
## Built With

* [![Go][go-shield]][go-url]
* [![Fiber][fiber-shield]][fiber-url]
* [![Fx][fx-shield]][fx-url]
* [![Redis][redis-shield]][redis-url]
* [![Prometheus][prometheus-shield]][prometheus-url]
* [![Swagger][swagger-shield]][swagger-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

* Go 1.25+ (for source builds)
  ```sh
  go version
  ```
* Docker (for the container workflow)
  ```sh
  docker --version
  ```
* A Redis instance (7.x), either local or containerized — see [Quick Start](#quick-start)

### Quick Start

Build the container image and run it alongside a Redis container:

```sh
make docker-build
docker run -d --name redis -p 6379:6379 redis:7
docker run -d --name monitor \
  -e REDIS__URL=redis://redis:6379 \
  --link redis \
  005-bot/monitor-go:$(git describe --tags --abbrev=0 2>/dev/null || echo 0.0.0)
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE -->
## Usage

### Build from Source

Run directly:

```sh
go run .
```

Or build and run the binary:

```sh
make build
./bin/monitor
```

### HTTP API

The service exposes an HTTP API (default bind `127.0.0.1:3000`):

* **Health probes** — `GET /health/live`, `GET /health/ready`, `GET /health/startup`
* **Swagger UI** — `GET /api/v1/docs`
* **Prometheus metrics** — `GET /metrics`
* **Pipeline endpoints** — storage (etag/diff/commit), scraper, parser, and monitor (status/run)

Full request examples for every endpoint are available in [`requests.http`](requests.http).

### Make Targets

Common development and quality targets:

```sh
make air     # run with live reload (air)
make test    # run tests with race detector and coverage
make lint    # run golangci-lint
make fmt     # format and generate code
make build   # build the binary into ./bin
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONFIGURATION -->
## Configuration

Configuration is loaded from environment variables using the `__` separator (e.g. `REDIS__URL`), with optional merge from a YAML file via `CONFIG_PATH`.

| Variable                  | Default                             | Description                          |
| ------------------------- | ----------------------------------- | ------------------------------------ |
| `REDIS__URL`              | `redis://localhost:6379`            | Redis connection URL                 |
| `SCRAPER__URL`            | `http://93.92.65.26/aspx/Gorod.htm` | HTML table source URL                |
| `SCRAPER__INTERVAL`       | `300`                               | Polling interval in seconds          |
| `STORAGE__TTL_DAYS`       | `5`                                 | Record retention in days             |
| `STORAGE__PREFIX`         | `bot-005`                           | Redis key prefix                     |
| `PUBLISHER__PREFIX`       | `bot-005`                           | Redis channel prefix                 |
| `PARSER__ADDRESS_DB_PATH` | (embedded `streets.db`)             | Custom SQLite database path          |
| `HTTP__ADDRESS`           | `127.0.0.1:3000`                    | HTTP server bind address             |
| `CONFIG_PATH`             | _(unset)_                           | Path to an optional YAML config file |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ROADMAP -->
## Roadmap

- [ ] Prometheus alerts based on scheduler error metrics
- [ ] Graceful shutdown improvements and circuit breaker on scrape failures
- [ ] Structured logging with request IDs across the pipeline

See the [open issues](https://github.com/005-bot/monitor-go/issues) for a full list of proposed features and known issues.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->
## Contributing

Contributions make the open-source community great. If you have a suggestion, please fork the repo and create a pull request. You can also open an issue with the "enhancement" tag.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->
## Contact

Maintainer: [@capcom6](https://github.com/capcom6)

Project Link: [https://github.com/005-bot/monitor-go](https://github.com/005-bot/monitor-go)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* [Go Fiber](https://github.com/gofiber/fiber)
* [Uber Fx](https://github.com/uber-go/fx)
* [go-redis](https://github.com/redis/go-redis)
* [miniredis](https://github.com/alicebob/miniredis)
* [goquery](https://github.com/PuerkitoBio/goquery)
* [Prometheus Client](https://github.com/prometheus/client_golang)
* [go-edlib](https://github.com/hbollon/go-edlib)
* [modernc.org/sqlite](https://gitlab.com/cznic/sqlite)
* [Koanf](https://github.com/knadh/koanf)
* [Best-README-Template](https://github.com/othneildrew/Best-README-Template)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
[contributors-shield]: https://img.shields.io/github/contributors/005-bot/monitor-go.svg?style=for-the-badge
[contributors-url]: https://github.com/005-bot/monitor-go/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/005-bot/monitor-go.svg?style=for-the-badge
[forks-url]: https://github.com/005-bot/monitor-go/network/members
[stars-shield]: https://img.shields.io/github/stars/005-bot/monitor-go.svg?style=for-the-badge
[stars-url]: https://github.com/005-bot/monitor-go/stargazers
[issues-shield]: https://img.shields.io/github/issues/005-bot/monitor-go.svg?style=for-the-badge
[issues-url]: https://github.com/005-bot/monitor-go/issues
[license-shield]: https://img.shields.io/github/license/005-bot/monitor-go.svg?style=for-the-badge
[license-url]: https://github.com/005-bot/monitor-go/blob/master/LICENSE
[go-shield]: https://img.shields.io/badge/go-1.25%2B-00ADD8?style=for-the-badge&logo=go
[go-url]: https://go.dev/
[fiber-shield]: https://img.shields.io/badge/Fiber-v2-00b894?style=for-the-badge&logo=go
[fiber-url]: https://github.com/gofiber/fiber
[fx-shield]: https://img.shields.io/badge/Uber%20Fx-DI-6f42c1?style=for-the-badge&logo=go
[fx-url]: https://github.com/uber-go/fx
[redis-shield]: https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis
[redis-url]: https://redis.io/
[prometheus-shield]: https://img.shields.io/badge/Prometheus-client-E6522C?style=for-the-badge&logo=prometheus
[prometheus-url]: https://prometheus.io/
[swagger-shield]: https://img.shields.io/badge/OpenAPI-Swagger-85EA2D?style=for-the-badge&logo=swagger
[swagger-url]: https://github.com/swaggo/swag
