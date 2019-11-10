FROM golang:alpine as builder
RUN apk update && \
    apk upgrade && \
    apk add git
RUN mkdir -p /go/src/botServer
ADD . /go/src/botServer
WORKDIR /go/src/botServer
RUN go get -u github.com/google/logger
RUN go get -u github.com/google/uuid
RUN go get -u github.com/pkg/errors
RUN go get -u github.com/gorilla/mux
RUN go build -o botServer .
FROM alpine
COPY --from=builder /go/src/botServer/botServer /app/
WORKDIR /app
CMD ["./botServer"]