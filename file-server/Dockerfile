FROM golang:1.20

ADD . /go/src/github.com/amruth-s05/eureka/file-server

WORKDIR /go/src/github.com/amruth-s05/file-server

RUN go install github.com/amruth-s05/file-server

ENTRYPOINT /go/bin/file-server