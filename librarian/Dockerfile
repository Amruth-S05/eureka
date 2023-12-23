FROM golang:1.9.1
ADD . /go/src/github.com/amruth-s05/eureka/librarian
WORKDIR /go/src/github.com/amruth-s05/eureka/librarian
RUN go install github.com/amruth-s05/eureka/librarian
ENTRYPOINT /go/bin/librarian
EXPOSE 9000