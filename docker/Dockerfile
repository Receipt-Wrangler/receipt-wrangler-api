FROM node:lts-alpine as node
WORKDIR /app

# Define build arguments
ARG VERSION
ARG BUILD_DATE

# Copy desktop source from local monolith
COPY desktop/ ./receipt-wrangler-desktop/

# Setup Desktop
RUN npm install -g @angular/cli
WORKDIR /app/receipt-wrangler-desktop

RUN npm install
RUN npm run build

# Setup API
FROM golang:1.26-trixie

# Define build arguments
ARG VERSION
ARG BUILD_DATE

# Copy API source from local monolith
COPY api/ /app/receipt-wrangler-api/

# Setup API
WORKDIR /app/receipt-wrangler-api

# Add local bin to path for python dependencies
ENV PATH="~/.local/bin:${PATH}"

# Set env
ENV ENV="prod"

# Set base path
ENV BASE_PATH="/app/receipt-wrangler-api"

# Set build date
ENV BUILD_DATE=${BUILD_DATE}

# Set version
ENV VERSION=${VERSION}

# Install tesseract dependencies
RUN bash set-up-dependencies.sh

# Build api
RUN go build

# Set up data volume
RUN mkdir data
VOLUME /app/receipt-wrangler-api/data

# Set up temp directory
RUN mkdir temp

# Set up sqlite volume
VOLUME /app/receipt-wrangler-api/sqlite

# Add logs volume
RUN mkdir logs
VOLUME /app/receipt-wrangler-api/logs

# Set pythonpath
ENV PYTHONPATH="/app/receipt-wrangler-api/wranglervenv/lib/python3.11/site-packages"

# Setup nginx
WORKDIR /app/receipt-wrangler-api/docker
COPY docker/ .
COPY docker/entrypoint.sh /app
RUN chmod 777 /app/entrypoint.sh

# Install nginx
RUN apt update
RUN apt install nginx -y

# Remove default configs
RUN rm /etc/nginx/sites-enabled/*
RUN rm /etc/nginx/sites-available/*
COPY docker/default.conf /etc/nginx/conf.d/default.conf

# Copy desktop dist
COPY --from=node /app/receipt-wrangler-desktop/dist/receipt-wrangler/browser /usr/share/nginx/html

# Clean up
WORKDIR /app

# Set up entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]

# Expose http port
EXPOSE 80
