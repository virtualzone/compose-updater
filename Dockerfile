FROM golang:1-alpine

RUN apk add docker py-pip python-dev libffi-dev openssl-dev gcc libc-dev make
RUN pip install docker-compose

WORKDIR /go/src/app
COPY *.go /go/src/app/

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]