# chat_app_backend

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Test Coverage](https://img.shields.io/badge/coverage-~40%25-yellow)](./docs/TEST_COVERAGE_SUMMARY.md)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](.)

**Live Demo: [https://chat-app.liu-yucheng.com/](https://chat-app.liu-yucheng.com/)**

> 🌍 **Languages:** English | [繁體中文](README_zh-TW.md)

## Overview

This project is a real-time chat application backend inspired by Discord architecture. It supports Servers (Guilds), Channels, Direct Messages (DM), Friends System, and File Uploads. The backend is built with Go, uses MongoDB for data storage, and integrates Redis for caching and session management.

## Key Features

- **Real-time Communication**: WebSocket-based instant messaging.
- **Microservices-ready Architecture**: Modular Controller → Service → Repository pattern.
- **Authentication**: JWT (Access/Refresh Token) and strict CSRF protection.
- **Testing Capabilities**: Highly unit-tested with a centralized mock schema (Overall Coverage ~40%, Middleware 94+%).
- **Deployment**: Supports Docker, Kubernetes (K3s/OrbStack), and GitOps ready configuration.
- **Load Testing**: Integrated K6 smoke and capacity testing scripts.

## Quick Start

### Prerequisites
Make sure you have Docker, Docker Compose, and `make` installed on your machine.

### Local Development

```bash
# 1. Initialize project dependencies
make init

# 2. Setup environment variables
cp .env.example .env.development

# 3. Start development environment (Docker Compose)
make dev

# 4. Run tests
make test
```

## Available Scripts

We provide a comprehensive `Makefile` to simplify local development:

- `make dev` - Start the development environment.
- `make logs` - Follow container logs.
- `make test` - Run unit tests.
- `make test-smoke` - Run a quick k6 smoke test.
- `make k8s-deploy` - Deploy to your local Kubernetes cluster (OrbStack/Minikube).
- `make k8s-delete` - Remove local Kubernetes deployments.

*For a complete list of commands, run `make help`.*

## Tech Stack

- **Language:** Go 1.23+
- **Framework:** Gin Web Framework
- **WebSocket:** gorilla/websocket
- **Database:** MongoDB
- **Cache:** Redis
- **Containerization:** Docker & Kubernetes (Kustomize)
