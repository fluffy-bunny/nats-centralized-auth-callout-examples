package decentralized_dynamic

import (
	"fmt"
	cobra_utils "natsauth/internal/cobra_utils"
	shared "natsauth/internal/shared"

	zerolog "github.com/rs/zerolog"
	cobra "github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const use = "decentralized_dynamic"

var (
	appInputs = shared.NewInputs()
)

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

			nc, err := appInputs.MakeConn(ctx)
			if err != nil {
				log.Error().Err(err).Msg("failed to connect to nats server")
				return err
			}
			defer nc.Drain()

			printer.Infof("%s connected to %s", appInputs.NatsUser, nc.ConnectedUrl())

			return nil
		},
	}
	appInputs.NatsUser = "alice"
	appInputs.NatsPass = "alice"
	appInputs.SentinelCreds = ""
	shared.InitCommonConnFlags(appInputs, command)

	flagName := "sentinel.creds"
	defaultS := appInputs.SentinelCreds
	command.Flags().StringVar(&appInputs.SentinelCreds, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	parentCmd.AddCommand(command)

}
