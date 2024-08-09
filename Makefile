# Makefile
protoc:
	cd pb && protoc -I . --grpc-gateway-ts_out . --go_out . --go_opt paths=source_relative --go-grpc_out . --go-grpc_opt paths=source_relative --grpc-gateway_out . --grpc-gateway_opt paths=source_relative ./**/*.proto

protocjs:
	cd pb && protoc -I . --go_out . --go_opt paths=source_relative --go-grpc_out . --go-grpc_opt paths=source_relative --grpc-gateway_out . --grpc-gateway_opt paths=source_relative ./**/*.proto