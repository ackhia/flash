package main

import (
	"log"
	"os"

	"github.com/ackhia/flash/config"
	fcrypto "github.com/ackhia/flash/crypto"
	"github.com/ackhia/flash/node"
	"github.com/ackhia/flash/p2p"
	"github.com/ackhia/flash/ui"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/spf13/cobra"
)

var logFile *os.File

func main() {
	//golog.SetAllLoggers(golog.LevelInfo)

	var rootCmd = &cobra.Command{Use: "flash"}

	var genCmd = &cobra.Command{
		Use:   "gen [keyfile filename]",
		Short: "Create a key pair",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			priv, _ := fcrypto.CreateKeyPair()
			fcrypto.WritePrivateKey(args[0], priv)
		},
	}

	var startCmd = &cobra.Command{
		Use:   "start [keyfile filename]",
		Short: "Start the node",
		Args:  cobra.MinimumNArgs(1),
	}
	var port int
	startCmd.Flags().IntVarP(&port, "port", "p", 0, "Port to listen on")
	startCmd.Run = func(cmd *cobra.Command, args []string) {
		priv, err := fcrypto.ReadPrivateKey(args[0])
		if err != nil {
			log.Fatalf("Could not read file %s %v", args[0], err)
		}

		startNode(priv, port)
	}

	rootCmd.AddCommand(genCmd, startCmd)
	rootCmd.Execute()
}

func startNode(privKey crypto.PrivKey, port int) {
	setupLogging()

	host, err := p2p.MakeHost(&privKey, port)
	if err != nil {
		log.Fatalf("Could not make host %v", err)
	}

	const bootstrapFilename = "bootstrap.yaml"
	bs, err := config.ReadBootstrapPeers(bootstrapFilename)
	if err != nil {
		log.Fatalf("Could not read %s %v", bootstrapFilename, err)
	}

	const genesisFilename = "genesis.yaml"
	genesis, err := config.ReadGenesis(genesisFilename)
	if err != nil {
		log.Fatalf("Could not read %s %v", genesisFilename, err)
	}

	n := node.New(privKey, &host, genesis, bs)

	n.Start()

	ui.Show(n)
	logFile.Close()
}

func setupLogging() {
	logFile, err := os.OpenFile("flash.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	log.SetOutput(logFile)
}
