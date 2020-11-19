FROM golang:1.15-alpine

WORKDIR /go/src/app
COPY . .
RUN apk add ffmpeg && \
  go get -d -v ./... && \
  go install -v ./...

ENTRYPOINT ["app"]
