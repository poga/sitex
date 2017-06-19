package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
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
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Serving %s at %s\n", *dir, addr)
	server.Start(listener)
}
