# Pull in source code to container
FROM golang:1.19.5-alpine3.17
WORKDIR .
RUN mkdir api
WORKDIR api
COPY . .

# Pull in configs from volume
WORKDIR /
RUN mkdir config
VOLUME ./config
RUN cp -a ./config/. /go/api

# Build api
WORKDIR /go/api
RUN ls && pwd
RUN go build

# Run
RUN ./api --env prod
EXPOSE 8081
