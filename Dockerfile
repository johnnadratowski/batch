# EXAMPLE SHAMELESSLY TAKEN FROM: http://blog.golang.org/docker
FROM golang

ADD . /go/src/github.com/Unified/batch
WORKDIR /go/src/github.com/Unified/batch

ENV GOPATH /go/src/github.com/Unified/batch/Godeps/_workspace:$GOPATH
RUN go build
RUN go install

# Run the batch command by default when the container starts.
ENTRYPOINT ["/go/bin/batch"]
