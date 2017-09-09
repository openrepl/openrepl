FROM golang:1.9-alpine

RUN apk add --no-cache git docker bash curl
RUN go get github.com/yudai/gotty github.com/yhat/wsutil

ADD filter.sh filter.sh
ADD sess.sh sess.sh
ADD server.go server.go

RUN go build -o /bin/server server.go
ENTRYPOINT ["/bin/server"]
