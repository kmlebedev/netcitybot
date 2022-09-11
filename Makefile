all: gen

.PHONY : gen

gen: dev

build:
	GOOS=linux CGO_ENABLED=0 go build -ldflags "-extldflags -static"

dev: build
	docker build --no-cache -t kmlebedev/netcitybot:local -f docker/Dockerfile.local .

go_build:
	docker build --no-cache -t kmlebedev/netcitybot:local -f docker/Dockerfile.go_build .

redis_up:
	docker-compose -f docker/docker-compose-redis-only.yml up

redis_down:
	docker-compose -f docker/docker-compose-redis-only.yml up
