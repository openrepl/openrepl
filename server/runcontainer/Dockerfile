FROM golang:1.9-alpine as builder
RUN apk add --no-cache git
RUN go get github.com/docker/docker/api github.com/jadr2ddude/websocket
COPY main.go main.go
RUN CGO_ENABLED=0 go build -o /runcontainer.o

FROM scratch
COPY --from=builder /runcontainer.o /bin/runcontainer
COPY langs.json langs.json
ENTRYPOINT ["/bin/runcontainer"]
