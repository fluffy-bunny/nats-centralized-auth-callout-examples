package jetstream

import (
	cobra_utils "natsauth/internal/cobra_utils"

	clients_jetstream_create "natsauth/cmd/cli/root/jetstream/create"

	cobra "github.com/spf13/cobra"
)

const use = "jetstream"

// Init command
func Init(parentCmd *cobra.Command) {
	var command = &cobra.Command{
		Use:               use,
		Short:             use,
		PersistentPreRunE: cobra_utils.ParentPersistentPreRunE,
	}

	clients_jetstream_create.Init(command)

	parentCmd.AddCommand(command)

}
