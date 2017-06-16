package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	dir := flag.String("dir", getWD(), "directory path")
	port := flag.Int("port", 8080, "port to use")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	server, err := NewServer(*dir, addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Serving %s at %s\n", *dir, addr)
	server.Start()
}

func getWD() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
}
