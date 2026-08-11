.PHONY: api config wire test

api:
	protoc --proto_path=./api \
		--proto_path=../../third_party \
		--go_out=paths=source_relative:./api \
		--go-grpc_out=paths=source_relative:./api \
		--go-http_out=paths=source_relative:./api \
		api/trip/v1/trip.proto

config:
	protoc --proto_path=./internal \
		--go_out=paths=source_relative:./internal \
		internal/conf/conf.proto

wire:
	cd cmd/trip && go run github.com/google/wire/cmd/wire

test:
	go test ./...
