package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net"
	"os"
)

func main() {
	// charset should have no effect at this case.
	// However, Chrome can't guess the right charset if there's no charset specified.
	// so we just default to charset=utf8
	mime.AddExtensionType(".json", "application/json; charset=utf8")
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dir := flag.String("dir", wd, "directory path")
	port := flag.Int("port", 8080, "port to use")
	flag.Parse()

	server, err := NewServer(*dir)
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", addr)

	fmt.Printf("Serving %s at %s\n", *dir, addr)
	server.Start(listener)
}
