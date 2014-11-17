# go.tftp

This is a work-in-progress TFTP implementation in Go.

## Installation

```
go get github.com/brooksbp/go.tftp
```

## Usage

Build the example server:

```
cd $GOPATH/src/github.com/brooksbp/go.tftp/cmd/tftp-server
go build
```

Try it out using a local TFTP client:

```
./tftp-server -listen 127.0.0.1:6969 &

# Generate a temporary 128K file and store it on the server. Fetch it back into a different filename and compare the contents.

dd if=/dev/urandom of=tmp.orig bs=1024 count=128
tftp -v 127.0.0.1 6969 -m binary -c put tmp.orig
tftp -v 127.0.0.1 6969 -m binary -c get tmp.orig tmp.recv
diff tmp.recv tmp.orig

fg
Ctrl-C
```
