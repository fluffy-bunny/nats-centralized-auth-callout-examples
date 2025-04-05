package clients

import (
	cobra_utils "natsauth/internal/cobra_utils"

	clients_decentralized_dynamic "natsauth/cmd/cli/root/clients/decentralized_dynamic"
	clients_micro "natsauth/cmd/cli/root/clients/micro"
	clients_request_reply "natsauth/cmd/cli/root/clients/request_reply"

	cobra "github.com/spf13/cobra"
)

const use = "clients"

// Init command
func Init(parentCmd *cobra.Command) {
	var command = &cobra.Command{
		Use:               use,
		Short:             use,
		PersistentPreRunE: cobra_utils.ParentPersistentPreRunE,
	}

	clients_request_reply.Init(command)
	clients_micro.Init(command)
	clients_decentralized_dynamic.Init(command)

	parentCmd.AddCommand(command)

}
