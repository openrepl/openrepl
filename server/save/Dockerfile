FROM golang:1.9-alpine

RUN apk --no-cache add git
RUN go get github.com/rs/cors
ADD save.go save.go
RUN go build -o /bin/save save.go
ENTRYPOINT ["/bin/save"]
