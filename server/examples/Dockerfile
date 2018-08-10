FROM golang:1.9-alpine as builder

COPY *.go /go/src/github.com/openrepl/server/examples/
COPY vendor /go/src/github.com/openrepl/server/examples/vendor
RUN CGO_ENABLED=0 go build -o /examples.o github.com/openrepl/server/examples

FROM scratch
COPY --from=builder /examples.o /bin/examples
COPY examples /examples
ENTRYPOINT ["/bin/examples"]
