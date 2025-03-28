package configurator

import (
	cobra_utils "natsauth/internal/cobra_utils"

	operator_mode "natsauth/cmd/cli/root/configurator/operator_mode"

	cobra "github.com/spf13/cobra"
)

const use = "configurator"

// Init command
func Init(parentCmd *cobra.Command) {
	var command = &cobra.Command{
		Use:               use,
		Short:             use,
		PersistentPreRunE: cobra_utils.ParentPersistentPreRunE,
	}

	operator_mode.Init(command)

	parentCmd.AddCommand(command)

}
