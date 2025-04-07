package services

import (
	cobra_utils "natsauth/internal/cobra_utils"

	callout_services_operator_mode_dynamic_accounts "natsauth/cmd/cli/root/callout/services/operator_mode_dynamic_accounts"
	operator_mode_url_resolver "natsauth/cmd/cli/root/callout/services/operator_mode_url_resolver"
	callout_services_static "natsauth/cmd/cli/root/callout/services/static"
	callout_services_url_resolver "natsauth/cmd/cli/root/callout/services/url_resolver"

	cobra "github.com/spf13/cobra"
)

const use = "services"

// Init command
func Init(parentCmd *cobra.Command) {
	var command = &cobra.Command{
		Use:               use,
		Short:             use,
		PersistentPreRunE: cobra_utils.ParentPersistentPreRunE,
	}

	callout_services_static.Init(command)
	callout_services_operator_mode_dynamic_accounts.Init(command)
	callout_services_url_resolver.Init(command)
	operator_mode_url_resolver.Init(command)

	parentCmd.AddCommand(command)

}
