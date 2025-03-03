package jetstream

import (
	clients_jetstream_consumer "natsauth/cmd/cli/root/jetstream/consumer"
	clients_jetstream_create "natsauth/cmd/cli/root/jetstream/create"
	clients_jetstream_info "natsauth/cmd/cli/root/jetstream/info"
	clients_jetstream_publish "natsauth/cmd/cli/root/jetstream/publish"
	cobra_utils "natsauth/internal/cobra_utils"

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
	clients_jetstream_info.Init(command)
	clients_jetstream_consumer.Init(command)
	clients_jetstream_publish.Init(command)

	parentCmd.AddCommand(command)

}
