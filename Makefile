dev-recreate:
	@docker-compose --project-name=rate-limiter-dev --env-file=deploy/dev/.env -f deploy/dev/docker-compose.yaml up -d --build --force-recreate

#dev-migration-up:
#	NETWORK_NAME=medical-chain-server-dev docker-compose --project-name=medical-chain-dev -f deploy/dev/docker-compose-migration-tool.yaml up up
#
#dev-migration-down:
#	NETWORK_NAME=medical-chain-server-dev docker-compose --project-name=medical-chain-dev -f deploy/dev/docker-compose-migration-tool.yaml up down
build-and-push-image: build-image push-image

build-image:
	@docker build . --target=release -t supermedicalchain/auth-service:pre-release

push-image:
	@docker tag supermedicalchain/auth-service:pre-release supermedicalchain/auth-service${TAG}
	@docker push supermedicalchain/auth-service${TAG}

build:
	@go build -o ./dist/server ./src

serve:
	@./dist/server serve

dev:
	@./dist/server --log-format plain --log-level debug --disable-profiler --allow-kill --remote-url http://test.medical.uetbc.xyz serve

cleanDB:
	@./dist/server clean

seed:
	@./dist/server seed-data --clean

test:
	#go test ./src/cockroach/... -v -check.f "CockroachDbGraphTestSuite.*"
	@go test ./... -v

test-prepare-up:
	@docker exec  up -f deploy/dev/docker-compose.yaml auth-cdb -d

test-prepare-down:
	@docker-compose down -f deploy/dev/docker-compose.yaml auth-cdb

grpc-client:
	@grpc-client-cli localhost:${GRPC_PORT}

kill:
	@(echo '{}' | grpc-client-cli -service CommonService -method Kill localhost:${GRPC_PORT}) > /nil 2> /nil || return 0
logs:
	@docker logs rate-limiter-dev_rate-limiter_1 -f
proto:
	@./genproto.sh