ARCH = amd64
.PHONY: all build_linux build_darwin build_windows

all: build_linux build_darwin build_windows

build_linux:
	./docker.sh linux $(arch)

build_darwin:
	./docker.sh darwin $(arch)

build_windows:
	./docker.sh windows $(arch)

clean:
	rm -f spyglass-*
