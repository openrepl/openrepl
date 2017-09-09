FROM golang:1.9-alpine

ADD store.go store.go
RUN go build -o /bin/store store.go
ENTRYPOINT ["/bin/store"]
