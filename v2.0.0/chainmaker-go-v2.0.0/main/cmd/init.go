/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"chainmaker.org/chainmaker-go/localconf"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
)

const (
	flagNameOfConfigFilepath          = "conf-file"
	flagNameShortHandOFConfigFilepath = "c"
)

func initLocalConfig(cmd *cobra.Command) {
	if err := localconf.InitLocalConfig(cmd); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func initFlagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	flags.StringVarP(&localconf.ConfigFilepath, flagNameOfConfigFilepath, flagNameShortHandOFConfigFilepath, localconf.ConfigFilepath, "specify config file path, if not set, default use ./chainmaker.yml")
	return flags
}

func attachFlags(cmd *cobra.Command, flagNames []string) {
	flags := initFlagSet()
	cmdFlags := cmd.Flags()
	for _, flagName := range flagNames {
		if flag := flags.Lookup(flagName); flag != nil {
			cmdFlags.AddFlag(flag)
		}
	}
}
