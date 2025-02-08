# Use Docker-in-Docker base image
FROM docker:23.0.6-dind

# Install common tools
RUN apk update && \
    apk add --no-cache curl bash git

# 1) Install kubectl
RUN curl -LO "https://dl.k8s.io/release/v1.26.3/bin/linux/amd64/kubectl" \
    && chmod +x kubectl \
    && mv kubectl /usr/local/bin/kubectl

# 2) Install kind
RUN curl -Lo kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64 \
    && chmod +x kind \
    && mv kind /usr/local/bin/kind

# 3) Install Go (example: Go 1.20.5)
RUN curl -LO https://go.dev/dl/go1.20.5.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.20.5.linux-amd64.tar.gz \
    && rm go1.20.5.linux-amd64.tar.gz

# Add Go to PATH
ENV PATH="/usr/local/go/bin:$PATH"

# By default, docker:dind runs dockerd in its entrypoint script.
CMD ["dockerd-entrypoint.sh"]
