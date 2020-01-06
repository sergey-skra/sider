SERVER_OUT := "bin/server"
CLIENT_OUT := "bin/client"
API_OUT := "pkg/pb/sider.pb.go"
API_REST_OUT := "pkg/pb/sider.pb.gw.go"
API_SWAGGER_OUT := "pkg/pb/sider.swagger.json"
PKG := "github.com/sergebraun/sider/"
SERVER_PKG_BUILD := "${PKG}/cmd/server"
CLIENT_PKG_BUILD := "${PKG}/cmd/client"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

.PHONY: all api server client pb rest swagger

all: api server client

pb: pkg/pb/sider.proto
	@protoc -I pkg/pb/ \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:pkg/pb \
		pkg/pb/sider.proto

rest: pkg/pb/sider.proto
	@protoc -I pkg/pb/ \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:pkg/pb \
		pkg/pb/sider.proto

swagger: pkg/pb/sider.proto
	@protoc -I pkg/pb/ \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--swagger_out=logtostderr=true:pkg/pb \
		pkg/pb/sider.proto

api: pb rest swagger ## Auto-generate grpc go sources

dep: ## Get the dependencies
	@go get -v -d ./...

server: dep api ## Build the binary file for server
	@go build -i -v -o $(SERVER_OUT) $(SERVER_PKG_BUILD)

client: dep api ## Build the binary file for client
	@go build -i -v -o $(CLIENT_OUT) $(CLIENT_PKG_BUILD)

clean: ## Remove previous builds
	@rm $(SERVER_OUT) $(CLIENT_OUT) $(API_OUT) $(API_REST_OUT) $(API_SWAGGER_OUT)

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
