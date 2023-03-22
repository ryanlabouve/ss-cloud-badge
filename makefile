.PHONY: build run

build:
	docker build -t ss-cloud-badge .

run: build
	docker run --rm -v "${PWD}/scans:/app/scans" ss-cloud-badge