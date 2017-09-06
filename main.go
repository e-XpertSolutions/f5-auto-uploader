package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
)

var configPath = flag.String("config", "config.toml", "path to configuration file")

func main() {
	flag.Parse()

	watchDir(nil, newLogger(os.Stderr), watchConfig{Dir: "/tmp/test"})

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Kill, os.Interrupt)

	<-sig

	fmt.Println("bye.")
}
