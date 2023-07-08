# Pull in source code to container
FROM golang:1.19.5-alpine3.17
WORKDIR .
RUN mkdir api
WORKDIR api
COPY . .

# Set up config volume
VOLUME /go/api/config

# Build api
WORKDIR /go/api
RUN go build

# Set up data volume
RUN mkdir data
VOLUME /go/api/data

# Run
RUN ./api --env prod
EXPOSE 8081
