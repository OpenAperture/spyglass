docker run --rm -it -e GOOS=$1 -e GOARCH=amd64 -v $GOPATH:/go -w /go/src/github.com/torrick/spyglass golang:1.4-cross ./build.sh
