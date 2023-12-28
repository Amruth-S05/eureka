FROM golang:1.20.11
ADD . /go/src/github.com/amruth-s05/eureka/doorman
WORKDIR /go/src/github.com/amruth-s05/eureka/doorman
RUN go install github.com/amruth-s05/eureka/doorman@latest
ENTRYPOINT /go/bin/doorman
EXPOSE 9000