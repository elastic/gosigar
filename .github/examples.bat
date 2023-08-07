@echo off
set examples=examples/df examples/free examples/ps examples/uptime
(for %%i in (%examples%) do (
    echo %%a
	go build -o %%i/out.exe ./%%i
	pushd %%i
	out.exe
	popd
))
