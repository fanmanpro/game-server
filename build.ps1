$env:GOOS="linux";$env:GOARCH="amd64"; go build ${workspaceRoot} -o builds/game-server-linux
$env:GOOS="windows";$env:GOARCH="amd64"; go build ${workspaceRoot} -o builds/game-server-windows.exe