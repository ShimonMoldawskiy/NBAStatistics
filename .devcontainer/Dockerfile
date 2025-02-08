# .devcontainer/Dockerfile
FROM mcr.microsoft.com/devcontainers/go:1.20

# Install kubectl
RUN apt-get update && apt-get install -y curl && \
    curl -sSLO "https://dl.k8s.io/release/v1.26.0/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && mv kubectl /usr/local/bin/kubectl
