package micro

import (
	"fmt"
	cobra_utils "natsauth/internal/cobra_utils"
	shared "natsauth/internal/shared"
	"time"

	zerolog "github.com/rs/zerolog"
	cobra "github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const use = "micro"

var (
	appInputs          = shared.NewInputs()
	requestData string = "hello"
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

			sub, _ := nc.SubscribeSync("rply")
			log.Info().Msgf("Sending request: %s", requestData)

			subject := "greet.joe"
			subLog := log.With().Str("subject", subject).Logger()
			srd := fmt.Sprintf("greet.joe: %s", requestData)
			err = nc.PublishRequest(subject, "rply", []byte(srd))
			if err != nil {
				subLog.Error().Err(err).Msg("failed to get response")
			} else {
				for start := time.Now(); time.Since(start) < 5*time.Second; {
					msg, err := sub.NextMsg(1 * time.Second)
					if err != nil {
						subLog.Error().Err(err).Msg("failed to get response")
						break
					}
					printer.Printf(cobra_utils.Blue, "Received: %s\n", string(msg.Data))
				}
			}

			subject = "greet.alice"
			subLog = log.With().Str("subject", subject).Logger()
			srd = fmt.Sprintf("greet.alice: %s", requestData)
			err = nc.PublishRequest(subject, "rply", []byte(srd))
			if err != nil {
				subLog.Error().Err(err).Msg("failed to get response")
			} else {
				for start := time.Now(); time.Since(start) < 5*time.Second; {
					msg, err := sub.NextMsg(1 * time.Second)
					if err != nil {
						subLog.Error().Err(err).Msg("failed to get response")
						break
					}
					printer.Printf(cobra_utils.Blue, "Received: %s\n", string(msg.Data))
				}
			}

			subject = "greet_junk.alice"
			subLog = log.With().Str("subject", subject).Logger()
			srd = fmt.Sprintf("greet_junk.alice: %s", requestData)
			err = nc.PublishRequest(subject, "rply", []byte(srd))
			if err != nil {
				subLog.Error().Err(err).Msg("failed to get response")
			} else {
				for start := time.Now(); time.Since(start) < 5*time.Second; {
					msg, err := sub.NextMsg(1 * time.Second)
					if err != nil {
						subLog.Error().Err(err).Msg("failed to get response")
						break
					}
					printer.Printf(cobra_utils.Blue, "Received: %s\n", string(msg.Data))
				}
			}

			return nil
		},
	}
	appInputs.NatsUser = "alice"
	appInputs.NatsPass = "alice"

	shared.InitCommonConnFlags(appInputs, command)

	flagName := "request.data"
	defaultS := requestData
	command.Flags().StringVar(&requestData, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	parentCmd.AddCommand(command)

}
