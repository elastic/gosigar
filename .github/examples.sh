#!/usr/bin/env bash

for exampleDir in "examples/df" "examples/free" "examples/ps" "examples/uptime" ; do
	go build -o ${exampleDir}/out.exe ./${exampleDir}
	pushd ${exampleDir}
	./out.exe
	popd
done
