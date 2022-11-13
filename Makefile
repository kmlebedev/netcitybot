all: gen

.PHONY : gen

gen: dev

build:
	GOOS=linux CGO_ENABLED=0 go build -ldflags "-extldflags -static"

dev: build
	docker build --no-cache -t kmlebedev/netcitybot:dev -f docker/Dockerfile.local .

dev_go_build: build
	docker build --no-cache -t kmlebedev/netcitybot:dev -f docker/Dockerfile.local_go_build .

go_build:
	docker build --no-cache -t kmlebedev/netcitybot:local -f docker/Dockerfile.go_build .

redis:
	docker-compose -f docker/docker-compose-redis-only.yml up

dev_redis: dev
	docker-compose --env-file .env -f docker/docker-compose-dev.yml up