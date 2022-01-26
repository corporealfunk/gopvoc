#!/bin/bash

version=$1

if [ "$version" == "" ]; then
  echo "Version number required"
  exit 1
fi

builds_dir="builds"
mkdir -p $builds_dir

build () {
  platform=$1
  arch=$2
  version=$3
  builds_dir=$4
  target_suffix=$5
  binary_extension=$6
  zip=$7
  target_name="gopvoc_${version}_${platform}_${arch}${target_suffix}"

  GOOS=$platform GOARCH=$arch go build -ldflags="-X 'main.Version=${version}'" -o "${builds_dir}/${target_name}/gopvoc${6}" main.go

  cd $builds_dir
  if [ "$zip" != "" ]; then
    zip -r "${target_name}.zip" ${target_name}
  else
    tar -zcvf "${target_name}.tar.gz" ${target_name}
  fi
  cd ..
}

build "darwin" "amd64" ${version} ${builds_dir} "_intel"
build "darwin" "arm64" ${version} ${builds_dir} "_m1"
build "linux" "amd64" ${version} ${builds_dir}
build "windows" "amd64" ${version} ${builds_dir} "" ".exe" 1
