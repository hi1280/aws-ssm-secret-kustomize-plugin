FROM golang:1.14 AS build

ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on

WORKDIR /go/src/github.com/hi1280/aws-ssm-secret-kustomize-plugin

COPY *.go go.mod go.sum ./

RUN go build -buildmode plugin -o AwsSsmSecret.so .
RUN go get sigs.k8s.io/kustomize/kustomize/v3

FROM ubuntu:xenial

WORKDIR /root
COPY --from=build /go/src/github.com/hi1280/aws-ssm-secret-kustomize-plugin/AwsSsmSecret.so .
COPY --from=build /go/bin/kustomize /usr/local/bin

CMD ["kustomize", "version"]