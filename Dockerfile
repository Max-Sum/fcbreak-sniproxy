FROM golang:1.19-alpine AS builder

WORKDIR /go/src/github.com/Max-Sum/fcbreak-sub

ARG GOPROXY=https://goproxy.io,direct

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build .

# Server Image
FROM scratch

WORKDIR /
COPY --from=builder \
     /go/src/github.com/Max-Sum/fcbreak-sub/fcbreak-sub /fcbreak-sub
COPY ./templates /templates

EXPOSE 8080
ENTRYPOINT [ "/fcbreak-sub" ]
CMD ["-s", "/service", "-l", ":8080"]