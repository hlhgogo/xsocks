export GOPROXY=https://goproxy.io,direct
export GOSUMDB=off

go_grpc_opt := --go_out . --go_opt paths=source_relative --go-grpc_out . --go-grpc_opt paths=source_relative
grpc_gw_opt := ${go_grpc_opt} --grpc-gateway_out . --grpc-gateway_opt paths=source_relative
ipaths      := -I. -I./proto/googleapis

prebuild:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.5.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.6.0

protobuf:
	protoc ${go_grpc_opt} ${ipaths} proto/xsocks/*.proto

.PHONY: spctl
spctl:
	go build -o output/spctl ./spctl
