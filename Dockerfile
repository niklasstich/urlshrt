FROM golang:1.15-alpine3.12 AS gobuild

WORKDIR /go/src/urlshrt
COPY . .

# First, get and run go-bindata again in case we dont have an up-to-date data.go
RUN go get -u -v github.com/go-bindata/go-bindata/...   && go-bindata -o data.go *.html

# Get dependencies and build project
RUN go get -d -v
RUN go build -o urlshrt

FROM alpine:3.12
COPY --from=gobuild /go/src/urlshrt/urlshrt /bin/urlshrt

CMD ["urlshrt"]