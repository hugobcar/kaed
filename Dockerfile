FROM golang:1.12.8 as builder
WORKDIR /go/src/github.com/hugobcar/kaed
ADD . /go/src/github.com/hugobcar/kaed
# RUN GO111MODULE=on go mod vendor
RUN CGO_ENABLED=0 go build -o kaed

FROM alpine:3.10.1
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /go/src/github.com/hugobcar/kaed/kaed .
CMD ["./kaed"]
