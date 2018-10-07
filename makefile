.PHONY: all clean new
.SILENT:

SRC=$(shell find src -name \*.go)

all: hike

clean:
	echo cleaning
	$(RM) hike

new: clean all

hike: $(SRC)
	echo building
	cd src; GOPATH="$(shell pwd)" go build -o ../hike
