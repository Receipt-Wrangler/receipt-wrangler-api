# syntax=docker/dockerfile:1
FROM golang:1.19.5-alpine3.17
WORKDIR .
RUN mkdir api
WORKDIR api
COPY . . 
RUN go build
RUN ./api
EXPOSE 8081


# Super basic example. We need the ability to create a config based on passed in env variables
