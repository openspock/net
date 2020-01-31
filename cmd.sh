#!/bin/sh

# build the protos
protoc -I=./proto --go_out=:$GOPATH/src ./proto/msg.proto

#openssl to generate key/cert pair