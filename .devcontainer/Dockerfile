FROM golang:1.23-bookworm

# Install all packages as root
RUN apt-get update && apt-get install -y git sudo zsh fzf curl \
    && rm -rf /var/lib/apt/lists/*

# Install golangci-lint (pinned version) system-wide for all users
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
    | sh -s -- -b /usr/local/bin v1.60.1

# Install zoxide
RUN curl -sS https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | bash

# Create non-root user with passwordless sudo and zsh shell
RUN useradd -m -s /bin/zsh vscode \
    && echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nopasswd \
    && chmod 0440 /etc/sudoers.d/nopasswd

# Set up Go environment for vscode user
ENV GOPATH=/home/vscode/.go
ENV GOMODCACHE=/home/vscode/.go/pkg/mod
ENV GOCACHE=/home/vscode/.cache/go-build
RUN mkdir -p /home/vscode/.go/pkg/mod /home/vscode/.cache/go-build \
    && chown -R vscode:vscode /home/vscode/.go /home/vscode/.cache

# Set ownership of workspace
WORKDIR /workspace
RUN chown -R vscode:vscode /workspace

USER vscode
