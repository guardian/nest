FROM golang:1.14

WORKDIR /go/src/app
COPY . .

RUN go get -d ./...
RUN go install ./...

ENTRYPOINT [ "nest", "http" ]