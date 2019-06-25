FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

#switch to our app directory
RUN mkdir -p /go/src/main
WORKDIR /go/src/main

#copy the source files
COPY main.go /go/src/main

#disable crosscompiling 
ENV CGO_ENABLED=0

#compile linux only
ENV GOOS=linux

RUN go get
#build the binary with debug information removed
RUN go build  -ldflags '-w -s' -a -installsuffix cgo -o main

FROM scratch

# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/main/main main

USER nobody
ENTRYPOINT ["./main"]