$env:GOOS="linux";$env:GOARCH="amd64"; go build ${workspaceRoot} -o game-server-linx
$env:GOOS="windows";$env:GOARCH="amd64"; go build ${workspaceRoot} -o game-server-windows.exe