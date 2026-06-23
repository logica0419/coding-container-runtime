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
	@
	mkdir -p rootfs
	docker export $$(docker create ubuntu) | tar -C rootfs -xvf -

.PHONY: clean
clean:
	@rm -f main
	@rm -rf rootfs

.PHONY: check
.ONESHELL: check
check:
	@
	echo "---------- 🔍セットアップ状況🔍 ----------"
	echo ""
	SETUP_COMPLETED=1

	if (uname >/dev/null 2>&1); then
		echo "[Linux] ✅ バージョン: $$(uname -o) $$(uname -r)"
	else
		SETUP_COMPLETED=0
		echo "[Linux] ❌ Linux上で実行して下さい"
	fi

	if [ "$$(stat -fc %T /sys/fs/cgroup/)" = "cgroup2fs" ]; then
		echo "[cgroup v2] ✅"
	else
		SETUP_COMPLETED=0
		echo "[cgroup2] ❌ cgroupがv2ではありません。OSをアップデートして下さい"
	fi

	if (type go >/dev/null 2>&1); then
		echo "[Go] ✅ バージョン: $$(go version)"
	else
		SETUP_COMPLETED=0
		echo "[Go] ❌ Goをインストールして下さい"
	fi

	if (type docker >/dev/null 2>&1); then
		echo "[Docker] ✅ バージョン: $$(docker -v)"
	else
		SETUP_COMPLETED=0
		echo "[Docker] ❌ Dockerをインストールして下さい"
	fi

	if [ -e ./config.json ]; then
		echo "[make init] ✅ ファイル生成済み"
	else
		SETUP_COMPLETED=0
		echo "[make init] ❌ \`make init\`を実行して下さい"
	fi

	if [ -d ./rootfs ]; then
		echo "[make rootfs] ✅ ファイル生成済み"
	else
		SETUP_COMPLETED=0
		echo "[make rootfs] ❌ \`make rootfs\`を実行して下さい"
	fi

	if [ $$SETUP_COMPLETED -eq 1 ]; then
		echo ""
		echo "---------- 🎉セットアップが完了しています🎉 ----------"
	else
		echo ""
		echo "---------- 引き続きセットアップを続けて下さい ----------"
	fi
