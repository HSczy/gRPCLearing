workDir = $(shell pwd)
grpcGenerate:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./route/route.proto

.PHONY: grpcGenerate