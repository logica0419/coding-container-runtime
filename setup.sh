#!/bin/sh -eu

repo_url=https://github.com/logica0419/coding-container-runtime.git
tmp_dir=$(mktemp -d)

trap 'rm -rf "$tmp_dir"' EXIT INT TERM HUP

for cmd in git make go docker; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "required command not found: $cmd" >&2
    exit 1
  fi
done

git clone --depth 1 --filter=blob:none --sparse "$repo_url" "$tmp_dir/repo"
git -C "$tmp_dir/repo" sparse-checkout set templates
cp -a "$tmp_dir/repo/templates/." .

make init
make rootfs
