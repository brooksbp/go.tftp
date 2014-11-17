package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/brooksbp/go.tftp"
)

var (
	flagListen = flag.String("listen", ":69", "host:port to listen on.")
)

func main() {
	flag.Parse()

	server, err := tftp.NewServer(*flagListen)
	if err != nil {
		fmt.Print(err)
	}

	_ = server.Run()

	os.Exit(0)
}
