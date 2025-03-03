package consume

import (
	"context"
	"fmt"
	cobra_utils "natsauth/internal/cobra_utils"
	shared "natsauth/internal/shared"
	"os"
	"os/signal"
	"syscall"

	fluffycore_async "github.com/fluffy-bunny/fluffycore/async"
	nats_jetstream "github.com/nats-io/nats.go/jetstream"
	async "github.com/reugn/async"
	zerolog "github.com/rs/zerolog"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

const use = "consume"

var (
	appInputs         = shared.NewInputs()
	appStreamConfig   = shared.NewStreamConfig()
	appConsumerConfig = nats_jetstream.ConsumerConfig{
		Name:           "",
		FilterSubjects: []string{},
	}
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
			ctx = log.WithContext(ctx)
			printer := cobra_utils.NewPrinter()
			printer.EnableColors = true

			ui := shared.NewUI(ctx)

			nc, err := appInputs.MakeConn(ctx)
			if err != nil {
				log.Error().Err(err).Msg("failed to connect to nats server")
				return err
			}
			defer nc.Drain()

			//printer.Infof("%s connected to %s", appInputs.NatsUser, nc.ConnectedUrl())

			js, err := nats_jetstream.New(nc)
			if err != nil {
				printer.Errorf("Error creating JetStream context: %v", err)
				return err
			}

			// get existing stream handle
			stream, err := js.Stream(ctx, appStreamConfig.Name)
			if err != nil {
				printer.Errorf("Error getting stream: %v", err)
				return err
			}
			// retrieve consumer handle from a stream
			consumer, err := stream.Consumer(ctx, appConsumerConfig.Name)
			if err != nil {
				printer.Errorf("Error getting consumer: %v", err)
				return err
			}

			ctxConsume, cancel := context.WithCancel(ctx)
			futurePublish := fluffycore_async.ExecuteWithPromiseAsync(func(promise async.Promise[*fluffycore_async.AsyncResponse]) {
				var err error
				defer func() {
					promise.Success(&fluffycore_async.AsyncResponse{
						Message: "End Serve - tview",
						Error:   err,
					})
				}()

				// consume messages from the consumer in callback
				cc, err := consumer.Consume(func(msg nats_jetstream.Msg) {
					subject := msg.Subject()

					ui.Main.Clear()

					mm := fmt.Sprintf("subject:%s message: %s", subject, string(msg.Data()))
					fmt.Fprintf(ui.Main, "%s ", mm)
					msg.Ack()
				})
				if err != nil {
					fmt.Fprint(ui.Main, err.Error())
					return
				}
				defer cc.Stop()

				quit := false
				for {
					if quit {
						break
					}
					select {
					case <-ctxConsume.Done():
						quit = true
					default:
					}

				}
			})

			// wait for an interrupt
			// Create a channel to receive OS signals.
			sigs := make(chan os.Signal, 1)
			// Notify the channel on interrupt signals.
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			fmt.Fprintf(ui.Footer, "%s ", "Waiting for interrupt signal...")

			fmt.Println("Waiting for interrupt signal...")

			// Block until a signal is received.
			<-sigs
			cancel()

			futurePublish.Join()
			ui.App.Stop()
			ui.Future.Join()
			//printer.Printf(cobra_utils.Green, "published %d messages\n", sequence+1)
			return nil

		},
	}
	appInputs.NatsUser = "god"
	appInputs.NatsPass = "god"

	shared.InitCommonConnFlags(appInputs, command)

	flagName := "js.name"
	defaultS := appStreamConfig.Name
	command.Flags().StringVar(&appStreamConfig.Name, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "consumer.name"
	defaultS = appConsumerConfig.Name
	command.Flags().StringVar(&appConsumerConfig.Name, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	parentCmd.AddCommand(command)

}
