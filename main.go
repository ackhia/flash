package main

import (
	"flag"

	"github.com/ackhia/flash/pkg/p2p"
	golog "github.com/ipfs/go-log/v2"
)

func main() {
	golog.SetAllLoggers(golog.LevelInfo) // Change to INFO for extra info

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	targetF := flag.String("d", "", "target peer to dial")
	insecureF := flag.Bool("insecure", false, "use an unencrypted connection")
	seedF := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	p2p.StartP2p(*listenF, *targetF, *insecureF, *seedF)
}
