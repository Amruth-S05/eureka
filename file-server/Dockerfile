FROM golang:1.20

ADD . /go/src/simplefs

WORKDIR /go/src/simlpefs

RUN go install simplefs

ENTRYPOINT /go/bin/simplefs