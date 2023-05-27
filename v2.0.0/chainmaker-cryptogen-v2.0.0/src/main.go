package main

import (
	"chainmaker.org/chainmaker-cryptogen/command"
	"chainmaker.org/chainmaker-cryptogen/config"
	"github.com/spf13/cobra"
)

func main() {
	mainCmd := &cobra.Command{
		Use: "chainmaker-cryptogen",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.LoadCryptoGenConfig(command.ConfigPath)
		},
	}
	mainFlags := mainCmd.PersistentFlags()
	mainFlags.StringVarP(&command.ConfigPath, "config", "c", "../config/crypto_config_template.yml", "specify config file path")

	mainCmd.AddCommand(command.ShowConfigCmd())
	mainCmd.AddCommand(command.GenerateCmd())
	mainCmd.AddCommand(command.ExtendCmd())

	mainCmd.Execute()

	return
}
