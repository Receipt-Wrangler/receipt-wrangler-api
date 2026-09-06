FROM golang:1.26-trixie

# First install curl if not already installed
RUN apt-get update
RUN apt-get install -y curl

# Setup NodeSource repository for the LTS version
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -

# Install Node.js
RUN apt-get install -y nodejs

# Install Java
RUN apt-get install -y default-jre

# Start in app dir
WORKDIR /app

# Copy repository into app directory
COPY . /app

# Setup Desktop
WORKDIR /app/desktop
RUN npm install -g @angular/cli
RUN npm install -g @openapitools/openapi-generator-cli

# Build desktop
RUN npm install
RUN npm run build

# Set up docs
WORKDIR /docs

# Clone docs source
RUN git clone https://github.com/Receipt-Wrangler/receipt-wrangler-doc.git
WORKDIR /docs/receipt-wrangler-doc

RUN npm install

# Setup API
WORKDIR /app/api

# Add local bin to path for python dependencies
ENV PATH="~/.local/bin:${PATH}"

# Set env
ENV ENV="dev"

# Set base path
ENV BASE_PATH="/app/api"

# Install tesseract dependencies
RUN bash set-up-dependencies.sh

# Install python dependencies for imap-client
RUN pip3 install -r imap-client/requirements.txt --break-system-packages

# Build api
RUN go build

# Set up data volume
RUN mkdir data
VOLUME /app/api/data

# Set up temp directory
RUN mkdir temp

# Set up sqlite volume
VOLUME /app/api/sqlite

# Add logs volume
RUN mkdir logs
VOLUME /app/api/logs

# Set pythonpath
ENV PYTHONPATH="/app/api/wranglervenv/lib/python3.11/site-packages"

# Expose http port
EXPOSE 80

# Setup Mobile (Flutter)
# Apt deps: unzip/zip + clang/cmake/ninja/gtk are Flutter's Linux desktop
# build toolchain; libsecret-1-dev is required to build the
# flutter_secure_storage_linux plugin (pkg-config fails CMake without it);
# xvfb lets the integration_test suite run headlessly under `xvfb-run`.
# See mobile/CLAUDE.md > E2E Testing for the full rationale.
RUN apt-get update && apt-get install -y --no-install-recommends \
      unzip zip clang cmake ninja-build libgtk-3-dev libsecret-1-dev xvfb \
    && rm -rf /var/lib/apt/lists/*

# Install Flutter SDK. Pin at image-build time; to bump, read the current
# stable hash from
# https://storage.googleapis.com/flutter_infra_release/releases/releases_linux.json
# (current_release.stable → releases[].version).
ENV FLUTTER_VERSION=3.41.7
RUN curl -fL "https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_${FLUTTER_VERSION}-stable.tar.xz" \
      -o /tmp/flutter.tar.xz \
    && tar xf /tmp/flutter.tar.xz -C /opt \
    && rm /tmp/flutter.tar.xz \
    && git config --global --add safe.directory /opt/flutter
ENV PATH="/opt/flutter/bin:${PATH}"

# Verify install, disable telemetry, enable Linux desktop target for e2e.
RUN flutter --version \
    && flutter config --no-analytics \
    && flutter config --enable-linux-desktop

# Warm the Linux-desktop engine cache and the mobile pub cache so the first
# `./run-e2e.sh` inside the container doesn't re-download ~500MB of artifacts.
RUN flutter precache --linux --no-ios --no-android --no-web --no-macos --no-windows
WORKDIR /app/mobile
RUN flutter pub get
WORKDIR /app

# Setup SSH
RUN apt-get update
RUN apt-get install -y curl openssh-server

# Set root password for SSH access (change 'your_password' to something secure)
RUN echo 'root:development' | chpasswd
# Configure SSH to allow root login
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
# Configure SSH to allow password authentication
RUN sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config

# Create startup script to start SSH
RUN echo '#!/bin/bash\n\
/usr/sbin/sshd\n\
tail -f /dev/null' > /app/startup.sh && \
chmod +x /app/startup.sh

# Expose SSH
EXPOSE 22

# Add Go and Flutter environment variables to root's bash profile
RUN echo 'export PATH=$PATH:/usr/local/go/bin:/go/bin' >> /root/.bashrc
RUN echo 'export GOPATH=/go' >> /root/.bashrc
RUN echo 'export PATH=$PATH:/opt/flutter/bin' >> /root/.bashrc

WORKDIR /app

# Keep er runnin
CMD ["/app/startup.sh"]
