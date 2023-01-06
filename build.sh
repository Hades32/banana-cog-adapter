# seems Banana's servers are quite old...
export GOAMD64=v3
export CGO_ENABLED=0
# ensure build is fully statically linked
go build -tags osusergo,netgo
