# FROM golang:alpine as builder
# RUN apk --no-cache add git build-base
# ENV HOME /go/src/todo
# WORKDIR $HOME
# COPY req.txt ./
# RUN cat req.txt | xargs go get
# COPY *.go ./
# RUN GOOS=linux go build -ldflags="-s -w"

FROM ubuntu:16.04
ENV HOME /app
WORKDIR $HOME
COPY todo $HOME/todo
# COPY --from=builder $HOME/todo .
# COPY front $HOME/front
# VOLUME $HOME/bin
EXPOSE 4000
ENTRYPOINT GIN_MODE=release $HOME/todo
