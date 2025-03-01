package shared

import (
	"context"
	"fmt"
	"os"

	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	status "github.com/gogo/status"
	nats "github.com/nats-io/nats.go"
	zerolog "github.com/rs/zerolog"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
	codes "google.golang.org/grpc/codes"
)

var _ctx context.Context

func SetContext(ctx context.Context) {
	_ctx = ctx
}
func GetContext() context.Context {
	return _ctx
}

type Inputs struct {
	NatsUrl         string   `json:"natsUrl"`
	NatsCreds       string   `json:"natsCreds"`
	IssuerSeed      string   `json:"issuerSeed"`
	NatsUser        string   `json:"natsUser"`
	NatsPass        string   `json:"natsPass"`
	XKeySeed        string   `json:"xkeySeed"`
	SigningKeyFiles []string `json:"signingKeyFiles"`
	UsersFile       string   `json:"usersFile"`
}

func NewInputs() *Inputs {
	return &Inputs{
		NatsUrl: "nats://localhost:4222",
	}
}

func InitCommonConnFlags(input *Inputs, command *cobra.Command) {
	flagName := "nats.url"
	defaultS := input.NatsUrl
	command.Flags().StringVar(&input.NatsUrl, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "nats.creds"
	defaultS = input.NatsCreds
	command.Flags().StringVar(&input.NatsCreds, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "nats.user"
	defaultS = input.NatsUser
	command.Flags().StringVar(&input.NatsUser, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "nats.pass"
	defaultS = input.NatsPass
	command.Flags().StringVar(&input.NatsPass, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

}

func (appInputs *Inputs) MakeConn(ctx context.Context) (*nats.Conn, error) {
	log := zerolog.Ctx(ctx).With().Str("command", "MakeConn").Logger()
	opts := []nats.Option{}
	if fluffycore_utils.IsNotEmptyOrNil(appInputs.NatsCreds) {
		if !FileExists(appInputs.NatsCreds) {
			log.Error().Msgf("nats creds file does not exist: %s", appInputs.NatsCreds)
			return nil, status.Error(codes.NotFound, fmt.Sprintf("nats creds file does not exist: %s", appInputs.NatsCreds))
		}
		opts = append(opts, nats.UserCredentials(appInputs.NatsCreds))
	}
	if fluffycore_utils.IsEmptyOrNil(appInputs.NatsUser) {
		log.Error().Msg("nats user is required")
		return nil, status.Error(codes.InvalidArgument, "nats user is required")
	}
	if fluffycore_utils.IsEmptyOrNil(appInputs.NatsPass) {
		log.Error().Msg("nats pass is required")
		return nil, status.Error(codes.InvalidArgument, "nats pass is required")
	}
	opts = append(opts, nats.UserInfo(appInputs.NatsUser, appInputs.NatsPass))

	nc, err := nats.Connect(
		appInputs.NatsUrl,
		opts...,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to nats server")
		return nil, err
	}
	return nc, nil

}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
