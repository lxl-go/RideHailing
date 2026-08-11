.PHONY: api wire test

api:
	protoc --proto_path=./api \
		--proto_path=../../third_party \
		--go_out=paths=source_relative:./api \
		--go-grpc_out=paths=source_relative:./api \
		--go-http_out=paths=source_relative:./api \
		api/auth/v1/auth.proto

wire:
	cd cmd/auth && go run github.com/google/wire/cmd/wire

test:
	go test ./...
