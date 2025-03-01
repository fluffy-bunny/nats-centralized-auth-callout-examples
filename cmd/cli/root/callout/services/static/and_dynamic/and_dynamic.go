package and_dynamic

import (
	cobra_utils "natsauth/internal/cobra_utils"
	shared "natsauth/internal/shared"

	zerolog "github.com/rs/zerolog"
	cobra "github.com/spf13/cobra"
)

const use = "and_dynamic"

// Init command
func Init(parentCmd *cobra.Command) {
	var command = &cobra.Command{
		Use:               use,
		Short:             use,
		PersistentPreRunE: cobra_utils.ParentPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := shared.GetContext()
			log := zerolog.Ctx(ctx).With().Str("command", use).Logger()

			printer := cobra_utils.NewPrinter()
			printer.EnableColors = true
			printer.PrintBold(cobra_utils.Bold, use)

			log.Info().Msg("TODO: and_dynamic")
			return nil
		},
	}

	parentCmd.AddCommand(command)

}
