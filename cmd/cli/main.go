/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"natsauth/cmd/cli/root"

	"github.com/rs/zerolog/log"
)

func main() {

	rootCommand := root.InitRootCmd()
	err := root.ExecuteE(rootCommand)
	if err != nil {
		log.Error().Err(err).Msg("error executing command")
	}
}
