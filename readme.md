## NBA Statistics
- A scalable system for NBA player's statistics
- The system runs on GitHub Codespaces
- Code written entirely in Golang
- Project delivered using Kubernetes
- A relational PostgreSQL database is used

## Application Architecture

### Load Balancer
A single entry point [https://laughing-memory-x5wxvr5rgpv529wv-8080.app.github.dev/](https://laughing-memory-x5wxvr5rgpv529wv-8080.app.github.dev/) that distributes incoming HTTP requests across multiple instances of your Go application.

### Golang Application
- Packaged and deployed in Docker containers; runs in several pods
- Uses connection pooling to PostgreSQL
- Integrates with a Redis caching layer

### Caching Layer (Redis)
- Stores frequently accessed or recently computed average values to reduce database load
- Ensures quick reads without always hitting the DB for the same queries

### PostgreSQL Database
- Primary store for records
- A Goose migration tool is used to handle schema changes

### Orchestration & Deployment
- Containers are orchestrated via Kubernetes
- Multiple replicas of the app run behind the load balancer
- Rolling updates for zero-downtime deployments: Kubernetes will start new pods with the updated version, then gracefully shut down old pods once the new ones pass health checkspplication Architecture

## High-Level Architecture Diagram
                      ┌───────────────────────┐
                      │  Client / Third-Party │
                      │    (e.g., other apps) │
                      └─────────┬─────────────┘
                                │
                      ┌─────────▼─────────┐
                      │   Load Balancer   │
                      │                   |
                      └─────────┬─────────┘
                                │
               ┌───────────────────────────────────────┐
               │               Kubernetes              │
               │                                       │
               └───────────────────────────────────────┘
                                │
                ┌───────────────┴────────────────┐
                │                                │
       ┌────────▼────────┐              ┌────────▼────────┐
       │ Go App Pod      │              │ Go App Pod      │
       │ (Container)     │              │ (Container)     │
       └────────┬────────┘              └────────┬────────┘
                │                                │
                │              ┌─────────────────┴───────────────────┐
                │              │          Caching Layer (Redis)      │
                └─────────────►│   (Caching aggregated query results)│
                               └─────────────────────────────────────┘
                                         │
                                         │
                               ┌───────────────────┐
                               │PostgreSQL Cluster │
                               │                   │
                               └───────────────────┘
