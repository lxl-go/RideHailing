.PHONY: api wire test

api:
	protoc --proto_path=./api \
		--proto_path=../../third_party \
		--go_out=paths=source_relative:./api \
		--go-grpc_out=paths=source_relative:./api \
		--go-http_out=paths=source_relative:./api \
		api/review/v1/review.proto

wire:
	cd cmd/review && go run github.com/google/wire/cmd/wire

test:
	go test ./...
