FROM golang

ENV PROJNAME=library_api
RUN mkdir -p /app/src/$PROJNAME
ENV GOPATH=/app
ENV GOBIN=$GOPATH/bin

COPY . /app/src/$PROJNAME


WORKDIR /app

RUN go-wrapper download $PROJNAME/main

RUN go build -o $GOBIN/server $PROJNAME/main


CMD $GOBIN/server

EXPOSE 2020

