# Applications

run:
	@echo "Choose run options:"
	@echo "[1] > User Gateway"
	@echo "[2] > User Service"
	@echo "[3] > Cancel"
	@read -p "Press [Enter]: " choice; \
    case "$$choice" in \
        1) make user-gateway;; \
        2) make user-service;; \
        3) exit 1;; \
        *) echo "Not Found Option"; exit 1;; \
    esac

user-gateway:
	nodemon --exec "go run" ./cmd/user-gateway/main.go --signal SIGTERM

user-service:
	nodemon --exec "go run" ./cmd/user-service/main.go --signal SIGTERM


# gRPC
gen-common:
	@protoc --go_out=. \
					--experimental_allow_proto3_optional \
					--go_opt=paths=source_relative \
					--go-grpc_out=. \
					--go-grpc_opt=paths=source_relative \
          proto/common/*.proto

gen-auth:
	@protoc --go_out=. \
					--experimental_allow_proto3_optional \
					--go_opt=paths=source_relative \
					--go-grpc_out=. \
					--go-grpc_opt=paths=source_relative \
          proto/auth-service/**/*.proto

gen-test:
	@protoc --go_out=. \
					--experimental_allow_proto3_optional \
					--go_opt=paths=source_relative \
					--go-grpc_out=. \
					--go-grpc_opt=paths=source_relative \
          proto/*.proto

gen-all: gen-common gen-auth