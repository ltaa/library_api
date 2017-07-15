FROM golang

COPY server /server

#ENTRYPOINT /server

WORKDIR /
CMD ./server > server.log


EXPOSE 2020

