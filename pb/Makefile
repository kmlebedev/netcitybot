all: gen

.PHONY : gen

gen:
	protoc netcity.proto --go_out=./netcity --go-grpc_out=./netcity --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative
