FROM golang:1.9-alpine as builder
COPY main.go /go/src/github.com/openrepl/server/store/main.go
RUN CGO_ENABLED=0 go build -o /store.o github.com/openrepl/server/store

FROM scratch
COPY --from=builder /store.o /bin/store
ENTRYPOINT ["/bin/store"]
