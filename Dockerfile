# Pull in source code to container
FROM golang:1.23.2-bullseye
WORKDIR .
RUN mkdir api
WORKDIR api
COPY . .

# Add local bin to path for python dependencies
ENV PATH="~/.local/bin:${PATH}"

# Set env
ENV ENV="prod"

# Set base path
ENV BASE_PATH="/go/api"

# Install tesseract dependencies
RUN ./set-up-dependencies.sh

# Build api
WORKDIR /go/api
RUN go build

# Set up data volume
RUN mkdir data
VOLUME /go/api/data

# Set up temp directory
RUN mkdir temp

# Set up sqlite volume
RUN mkdir sqlite
VOLUME /go/api/sqlite

# Add logs volume
RUN mkdir logs
VOLUME /go/api/logs

# Expose port
EXPOSE 8081
