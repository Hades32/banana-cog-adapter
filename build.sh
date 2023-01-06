# seems Banana's servers are quite old...
export GOAMD64=v3
export CGO_ENABLED=0
go build -tags osusergo,netgo
