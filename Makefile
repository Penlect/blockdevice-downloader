
all: image

overlays/blockdevice-downloader/usr/local/bin/blockdevice-downloader: main.go
	go build -o overlays/blockdevice-downloader/usr/local/bin/blockdevice-downloader main.go

image: debos.yaml overlays/blockdevice-downloader/usr/local/bin/blockdevice-downloader
	debos -c 8 -m 4GB --print-recipe -t buildbase:$(shell [ -f base.tar.gz~ ] && echo no || echo yes) debos.yaml

clean:
	-rm base.tar.gz~ *.img *.img.gz *.img.bmap
