FROM golang:1.9-alpine
RUN apk add --no-cache git
RUN go get -u github.com/motemen/gore github.com/nsf/gocode github.com/k0kubun/pp github.com/davecgh/go-spew/spew golang.org/x/tools/cmd/godoc

ADD rungo.sh rungo.sh
ENTRYPOINT ["sh", "rungo.sh"]
