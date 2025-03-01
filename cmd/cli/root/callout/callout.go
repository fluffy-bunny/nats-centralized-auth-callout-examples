package callout

import (
	cobra_utils "natsauth/internal/cobra_utils"

	callout_services "natsauth/cmd/cli/root/callout/services"

	cobra "github.com/spf13/cobra"
)

const use = "callout"

// Init command
func Init(parentCmd *cobra.Command) {
	var command = &cobra.Command{
		Use:               use,
		Short:             use,
		PersistentPreRunE: cobra_utils.ParentPersistentPreRunE,
	}

	callout_services.Init(command)

	parentCmd.AddCommand(command)

}
