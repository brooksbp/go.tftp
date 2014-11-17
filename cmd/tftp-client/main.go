package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/brooksbp/go.tftp"
)

var (
	flagHelp = flag.Bool("h", false, "show this help")
	flagHost = flag.String("host", ":69", "host:port to connect to.")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if *flagHelp {
		usage()
	}

	client, err := tftp.NewClient(*flagHost)
	if err != nil {
		fmt.Print(err)
	}

	_ = client

	/*
		// Get a local address and setup a listening conn on it.
		laddr, err := net.ResolveUDPAddr("udp", ":0")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Get a remote 69 address and send an initial RRQ to it.
		raddr69, err := net.ResolveUDPAddr("udp", *flagHost)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		_, err = conn.WriteToUDP([]byte("RRQ"), raddr69)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Read a message on our local conn.. should be from a new
		// UDP port on the server.
		var buf [512]byte
		n, raddr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%s %s READ: %s\n",
			conn.LocalAddr(), raddr,
			string(buf[0:n]))
	*/
	os.Exit(0)
}
