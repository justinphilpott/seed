FROM mcr.microsoft.com/devcontainers/go:2-1.25-trixie

# Pre-create .vscode-server owned by vscode so VS Code's server setup
# (which runs before any lifecycle commands) can write bin/ and data/ inside it.
RUN mkdir -p /home/vscode/.vscode-server /home/vscode/.config \
    && chown vscode:vscode /home/vscode/.vscode-server /home/vscode/.config
