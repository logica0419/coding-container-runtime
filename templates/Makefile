.PHONY: init
init:
	@go mod download
	@sudo $$(which go) mod download
	@/bin/echo -e "{\n  \"name\": \"container\",\n  \"entry_point\": [\"bash\"],\
	\n  \"cgroup\": {\n    \"max_cpu_percent\": 100,\n    \"max_memory_mb\": 1024\n  },\
	\n  \"rootfs\": {\n    \"rootfs_path\": \"./rootfs\"\n  }\n}" > ./config.json

main: *.go
	go build -o main *.go

.PHONY: run
run: main
	./main run bash

.ONESHELL: rootfs
rootfs:
	@mkdir -p rootfs
	@docker export $$(docker create ubuntu) | tar -C rootfs -xvf -
	@cp ./dev/stress ./rootfs/usr/bin

.PHONY: clean
clean:
	@rm -f main
	@rm -rf rootfs
	@rm -rf spec

.PHONY: check
check:
	@sh ./dev/env_check.sh
