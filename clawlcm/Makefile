.PHONY: build test clean docker-build docker-run docker-stop podman-build podman-run podman-stop podman-compose-build podman-compose-up podman-compose-down

build:
	CGO_ENABLED=1 go build -o bin/clawlcm ./cmd/clawlcm

test:
	go test -v ./...

clean:
	rm -f clawlcm
	rm -f data/*.db

docker-build:
	docker build -t clawlcm:latest -f docker/Dockerfile .

docker-run:
	docker run -d --name clawlcm \
		-v clawlcm-data:/app/data \
		clawlcm:latest

docker-stop:
	docker stop clawlcm || true
	docker rm clawlcm || true

podman-build:
	podman build -t clawlcm:latest -f docker/Dockerfile .

podman-run:
	podman run -d --name clawlcm \
		-v clawlcm-data:/app/data \
		clawlcm:latest

podman-stop:
	podman stop clawlcm || true
	podman rm clawlcm || true

podman-compose-build:
	podman compose -f docker/docker-compose.yml build

podman-compose-up:
	podman compose -f docker/docker-compose.yml up -d

podman-compose-down:
	podman compose -f docker/docker-compose.yml down
