# go.tftp

This is a work-in-progress TFTP implementation in Go.

## Installation

```
go get github.com/brooksbp/go.tftp
```

## Usage

Example server:

```
cd $GOPATH/src/github.com/brooksbp/go.tftp/cmd/tftp-server
go build

./tftp-server -listen 127.0.0.1:6969 &

# Generate a temporary 128K file.

dd if=/dev/urandom of=tmp.orig bs=1024 count=128

# Use a local TFTP client to put & get the file.

tftp -v 127.0.0.1 6969 -m binary -c put tmp.orig
tftp -v 127.0.0.1 6969 -m binary -c get tmp.orig tmp.recv
diff tmp.recv tmp.orig

fg ; Ctrl-c
```
