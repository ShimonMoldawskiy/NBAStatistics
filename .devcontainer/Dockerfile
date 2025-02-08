FROM ubuntu:22.04

# Install common dev tools as needed
RUN apt-get update && apt-get install -y \
    curl git bash ca-certificates \
    && rm -rf /var/lib/apt/lists/*

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

# Just use a shell as the default command
CMD ["/bin/bash"]