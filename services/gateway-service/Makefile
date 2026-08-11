.PHONY: api test

api:
	protoc --proto_path=./api \
		--proto_path=D:/GoDev/Workspace/pkg/mod/github.com/go-kratos/kratos/v2@v2.9.2/third_party \
		--go_out=paths=source_relative:./api \
		--go-http_out=paths=source_relative:./api \
		api/gateway/v1/trip.proto

test:
	go test ./...
