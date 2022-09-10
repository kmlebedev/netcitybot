all: gen

.PHONY : gen

gen: dev

build:
	GOOS=linux CGO_ENABLED=0 go build -ldflags "-extldflags -static"

dev: build
	docker build --no-cache -t kmlebedev/netcitybot:local -f docker/Dockerfile.local .

go_build:
	docker build --no-cache -t kmlebedev/netcitybot:local -f docker/Dockerfile.go_build .
