FROM golang:1.16 AS build
#RUN apk update
#RUN mkdir "dirname"
#FROM alpine:3.12
WORKDIR /app

RUN go env -w GOPROXY=https://goproxy.cn,direct
COPY . .
RUN go mod tidy
RUN GOOS=linux CGO_ENABLED=1 GOARCH=amd64 go build -o dockermain
ENTRYPOINT ["./dockermain"]