set examples=examples/df examples/free examples/ps examples/uptime
(for %%i in (%examples%) do (
	go build -o %%i/out.exe ./%%i
	%%i/out.exe
))
