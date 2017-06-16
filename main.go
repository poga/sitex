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
	log.Fatal(Serve(*dir, fmt.Sprintf(":%d", *port)))
}

func getWD() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
}
