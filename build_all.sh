#!/bin/bash

version=$1

builds_dir="builds"
mkdir -p $builds_dir

# Darin amd64
darwin_amd64_dir="gopvoc_${version}_darwin_amd64_intel"

GOOS=darwin GOARCH=amd64 go build -o "${builds_dir}/${darwin_amd64_dir}/gopvoc" main.go

# Darin arm64
darwin_arm64_dir="gopvoc_${version}_darwin_arm64_m1"

GOOS=darwin GOARCH=arm64 go build -o "${builds_dir}/${darwin_arm64_dir}/gopvoc" main.go

# Linux amd64
linux_amd64_dir="gopvoc_${version}_linux_amd64"

GOOS=linux GOARCH=amd64 go build -o "${builds_dir}/${linux_amd64_dir}/gopvoc" main.go

# Windows amd64
windows_amd64_dir="gopvoc_${version}_windows_amd64"

GOOS=windows GOARCH=amd64 go build -o "${builds_dir}/${windows_amd64_dir}/gopvoc.exe" main.go

# tar and zip:
cd $builds_dir

tar -zcvf "${darwin_amd64_dir}.tar.gz" $darwin_amd64_dir
tar -zcvf "${darwin_arm64_dir}.tar.gz" $darwin_arm64_dir
tar -zcvf "${linux_amd64_dir}.tar.gz" $linux_amd64_dir
zip -r "${windows_amd64_dir}.zip" $windows_amd64_dir
