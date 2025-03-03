package consumer

import (
	clients_jetstream_consumer_add "natsauth/cmd/cli/root/jetstream/consumer/add"
	clients_jetstream_consumer_info "natsauth/cmd/cli/root/jetstream/consumer/info"
	cobra_utils "natsauth/internal/cobra_utils"

	cobra "github.com/spf13/cobra"
)

const use = "consumer"

// Init command
func Init(parentCmd *cobra.Command) {
	var command = &cobra.Command{
		Use:               use,
		Short:             use,
		PersistentPreRunE: cobra_utils.ParentPersistentPreRunE,
	}

	clients_jetstream_consumer_add.Init(command)
	clients_jetstream_consumer_info.Init(command)

	parentCmd.AddCommand(command)

}
