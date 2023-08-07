
for %%exampleDir in (examples/df examples/free examples/ps examples/uptime) do (
	go build -o %%exampleDir/out.exe ./%%exampleDir
	cd %%exampleDir
	out.exe
	cd ..
)
