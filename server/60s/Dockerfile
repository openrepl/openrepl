FROM golang:1.9-alpine

ADD store.go store.go
RUN apk --no-cache add git
RUN go get github.com/rs/cors
RUN go build -o /bin/store store.go
ENTRYPOINT ["/bin/store"]
