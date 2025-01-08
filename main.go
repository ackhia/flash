package main

import (
	"fmt"
	"log"

	golog "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
)

func main() {
	golog.SetAllLoggers(golog.LevelInfo) // Change to INFO for extra info

	// Parse options from the command line
	/*listenF := flag.Int("l", 0, "wait for incoming connections")
	targetF := flag.String("d", "", "target peer to dial")
	insecureF := flag.Bool("insecure", false, "use an unencrypted connection")
	seedF := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()*/

	var rootCmd = &cobra.Command{Use: "flash"}

	var genCmd = &cobra.Command{
		Use:   "gen [keyfile filename]",
		Short: "Create a key pair",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			priv, pub := CreateKeyPair()
			WriteKeyfile(args[0], priv, pub)
		},
	}

	var startCmd = &cobra.Command{
		Use:   "start [keyfile filename]",
		Short: "Start the node",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			priv, pub, err := ReadKeyfile(args[0])
			if err != nil {
				log.Fatalf("Could not read file %s %s", args[0], err)
			}

			privBytes, _ := priv.Raw()
			pubBytes, _ := pub.Raw()

			fmt.Printf("Private key: %s, Public key: %s", privBytes, pubBytes)
		},
	}

	rootCmd.AddCommand(genCmd, startCmd)
	rootCmd.Execute()

	//p2p.StartP2p(*listenF, *targetF, *insecureF, *seedF)
}
