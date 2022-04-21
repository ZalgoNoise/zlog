#!/bin/bash
protoc --go_out=. proto/event.proto 
protoc --go_out=plugins=grpc:proto proto/service.proto
sed -i 's|event "./log/event"|event "github.com/zalgonoise/zlog/log/event"|'  proto/service/service.pb.go 