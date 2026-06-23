#!/bin/sh -eu

repo_url=https://github.com/logica0419/coding-container-runtime.git
tmp_dir=$(mktemp -d)

trap 'rm -rf "$tmp_dir"' EXIT INT TERM HUP

git clone --depth 1 --filter=blob:none --sparse "$repo_url" "$tmp_dir/repo"
git -C "$tmp_dir/repo" sparse-checkout set templates
cp -a "$tmp_dir/repo/templates/." .
