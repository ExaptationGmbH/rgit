# .claude-container/Dockerfile
#
# Per-project image extension. claude-container detects this file and
# builds an image on top of the base claude-container image, then
# launches that image instead. The built image is cached and only
# rebuilt when either this file or the base image changes.
#
# RULES (enforced by claude-container):
#   - First instruction MUST be: FROM claude-container:latest
#   - Do NOT add ENTRYPOINT or CMD — claude-container manages startup.
#   - The last USER SHOULD be 'claude' (root works but triggers a warning).
#
# Anything else is fair game. Edit freely.

FROM claude-container:latest

# ---------------------------------------------------------------------
# OS packages (apt)
# ---------------------------------------------------------------------
# The base image is Debian (node:23-bookworm). Drop to root to install
# system packages, then switch back to the unprivileged user below.
#
# Uncomment and add the packages you need:
#
# USER root
# RUN apt-get update && apt-get install -y --no-install-recommends \
#     python3 \
#     build-essential \
#     && rm -rf /var/lib/apt/lists/*

# Go toolchain — needed to build and test the `rgit` utility that lives
# in this repo. Installed from the official tarball (Debian's golang-go
# lags behind), then put on PATH for the claude user.
USER root
RUN curl -fsSL https://go.dev/dl/go1.22.5.linux-arm64.tar.gz -o /tmp/go.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz
ENV PATH="/usr/local/go/bin:/home/claude/go/bin:${PATH}"

# ---------------------------------------------------------------------
# Language toolchains (user-scoped installs)
# ---------------------------------------------------------------------
# Run these as the 'claude' user so files land in /home/claude and are
# owned correctly at runtime. Uncomment what you need.

USER claude

# Rust (rustup):
# RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \
#     | sh -s -- -y --default-toolchain stable
# ENV PATH="/home/claude/.cargo/bin:${PATH}"

# uv (Python package & project manager):
# RUN curl -LsSf https://astral.sh/uv/install.sh | sh
# ENV PATH="/home/claude/.local/bin:${PATH}"

# Global npm packages (node is already in the base image):
# RUN npm install -g pnpm yarn typescript

# ---------------------------------------------------------------------
# Environment variables
# ---------------------------------------------------------------------
# ENV MY_TOOL_FLAG=1
# ENV PYTHONUNBUFFERED=1

# ---------------------------------------------------------------------
# Project-specific setup (optional)
# ---------------------------------------------------------------------
# The build context is .claude-container/, so any extra files you want
# COPY'd in must live there (NOT in the project root). Example:
#
# COPY setup.sh /tmp/setup.sh
# RUN bash /tmp/setup.sh && rm /tmp/setup.sh

# Leave the last USER as 'claude' so the container runs unprivileged.
USER claude
