FROM golang:1.9-alpine as builder
RUN apk add --no-cache git
COPY *.go /go/src/github.com/openrepl/server/runcontainer/
COPY vendor /go/src/github.com/openrepl/server/runcontainer/vendor
RUN CGO_ENABLED=0 go build -o /runcontainer.o github.com/openrepl/server/runcontainer

FROM scratch
COPY --from=builder /runcontainer.o /bin/runcontainer
COPY langs.json langs.json
ENTRYPOINT ["/bin/runcontainer"]
