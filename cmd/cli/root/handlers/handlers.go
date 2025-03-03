package handlers

import (
	cobra_utils "natsauth/internal/cobra_utils"

	handlers_micro "natsauth/cmd/cli/root/handlers/micro"
	handlers_request "natsauth/cmd/cli/root/handlers/request"

	cobra "github.com/spf13/cobra"
)

const use = "handlers"

// Init command
func Init(parentCmd *cobra.Command) {
	var command = &cobra.Command{
		Use:               use,
		Short:             use,
		PersistentPreRunE: cobra_utils.ParentPersistentPreRunE,
	}

	handlers_request.Init(command)
	handlers_micro.Init(command)

	parentCmd.AddCommand(command)

}
