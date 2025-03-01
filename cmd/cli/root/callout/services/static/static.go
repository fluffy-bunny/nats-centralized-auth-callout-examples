package static

import (
	"fmt"
	cobra_utils "natsauth/internal/cobra_utils"
	"os"
	"os/signal"
	"syscall"
	"time"

	shared "natsauth/internal/shared"

	callout_services_static_and_dynamic "natsauth/cmd/cli/root/callout/services/static/and_dynamic"

	callout "github.com/aricart/callout.go"
	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	status "github.com/gogo/status"
	jwt "github.com/nats-io/jwt/v2"
	nkeys "github.com/nats-io/nkeys"
	zerolog "github.com/rs/zerolog"
	cobra "github.com/spf13/cobra"
	"github.com/spf13/viper"
	codes "google.golang.org/grpc/codes"
)

const use = "static"

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

			// parse the private key
			akp, err := nkeys.FromSeed([]byte(appInputs.IssuerSeed))
			if err != nil {
				log.Error().Err(err).Msg("error parsing issuer seed")
				return err
			}
			akpPublickKey, _ := akp.PublicKey()
			log.Info().Str("issuer", akpPublickKey).Msg("issuer")

			// Parse the xkey seed if present.
			var curveKeyPair nkeys.KeyPair
			if fluffycore_utils.IsNotEmptyOrNil(appInputs.XKeySeed) {
				curveKeyPair, err = nkeys.FromSeed([]byte(appInputs.XKeySeed))
				if err != nil {
					log.Error().Err(err).Msg("error parsing xkey seed")
					return status.Error(codes.Internal, "error parsing xkey seed")
				}
			}
			if curveKeyPair != nil {
				curveKeyPairPublicKey, _ := curveKeyPair.PublicKey()
				log.Info().Str("xkey", curveKeyPairPublicKey).Msg("xkey")
			}
			// a function that creates the users
			authorizer := func(req *jwt.AuthorizationRequest) (string, error) {
				// peek at the req for information - for brevity
				// in the example, we simply allow them in
				log.Info().Str("user", req.UserNkey).Msg("authorizing")
				// use the server specified user nkey
				uc := jwt.NewUserClaims(req.UserNkey)
				// put the user in the global account
				uc.Audience = "SVC"
				// add whatever permissions you need
				uc.Sub.Allow.Add(">")
				// perhaps add an expiration to the JWT
				uc.Expires = time.Now().Unix() + 90
				return uc.Encode(akp)
			}
			// start the service
			_, err = callout.NewAuthorizationService(nc, callout.Authorizer(authorizer), callout.ResponseSignerKey(akp))
			if err != nil {
				log.Error().Err(err).Msg("error starting service")
				return err
			}
			// don't exit until sigterm
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
			<-quit
			return nil
		},
	}
	appInputs.NatsUser = "auth"
	appInputs.NatsPass = "auth"
	appInputs.IssuerSeed = "SAAEXFSYMLINXLKR2TG5FLHCJHLU62B3SK3ESZLGP4B4XGLUNXICW3LGAY"

	shared.InitCommonConnFlags(appInputs, command)

	flagName := "issuer.seed"
	defaultS := appInputs.IssuerSeed
	command.Flags().StringVar(&appInputs.IssuerSeed, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	callout_services_static_and_dynamic.Init(command)
	parentCmd.AddCommand(command)

}
