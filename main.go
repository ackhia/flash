package main

import (
	"log"
	"os"

	"github.com/ackhia/flash/config"
	fcrypto "github.com/ackhia/flash/crypto"
	"github.com/ackhia/flash/node"
	"github.com/ackhia/flash/p2p"
	"github.com/ackhia/flash/ui"
	golog "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/spf13/cobra"
)

func main() {
	golog.SetAllLoggers(golog.LevelInfo)

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
		Run: func(cmd *cobra.Command, args []string) {
			priv, err := fcrypto.ReadPrivateKey(args[0])
			if err != nil {
				log.Fatalf("Could not read file %s %v", args[0], err)
			}

			startNode(priv)
		},
	}

	rootCmd.AddCommand(genCmd, startCmd)
	rootCmd.Execute()
}

func startNode(privKey crypto.PrivKey) {
	//setupLogging()

	host, err := p2p.MakeHost(&privKey)
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
}

func setupLogging() {
	logFile, err := os.Create("log.txt")
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
}
