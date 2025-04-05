package operator_mode

import (
	"context"
	"fmt"
	configurator_models "natsauth/cmd/cli/root/configurator/models"
	configurator_shared "natsauth/cmd/cli/root/configurator/shared"
	cobra_utils "natsauth/internal/cobra_utils"
	"natsauth/internal/shared"

	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	jwxt "github.com/lestrrat-go/jwx/v2/jwt"
	zerolog "github.com/rs/zerolog"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const use = "operator_mode"

type (
	ResolverParams struct {
		Name string
	}
)

var (
	CurrentResolverParams = ResolverParams{}
)

func decodeJwt(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "decodeJwt").Logger()
	notTrustedToken, err := jwxt.ParseString(tokenString,
		jwxt.WithValidate(false),
		jwxt.WithVerify(false))
	if err != nil {
		log.Error().Err(err).Msg("failed to parse JWT")
		return nil, err
	}

	return notTrustedToken.AsMap(ctx)

}

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

			printer.Info(fluffycore_utils.PrettyJSON(CurrentResolverParams))

			if fluffycore_utils.IsEmptyOrNil(CurrentResolverParams.Name) {
				return status.Error(codes.InvalidArgument, "name is required")
			}
			createOperatorResponse, err := configurator_shared.CreateOperator(ctx,
				&configurator_models.CreateOperatorRequest{
					Name: CurrentResolverParams.Name,
				})
			if err != nil {
				log.Error().Err(err).Msg("failed to create operator")
				return err
			}
			printer.Info(fluffycore_utils.PrettyJSON(createOperatorResponse))

			jwtJson, err := decodeJwt(ctx, createOperatorResponse.OperatorAccount.JWT)
			if err != nil {
				log.Error().Err(err).Msg("failed to decode operator jwt")
				return err
			}
			printer.Info(fluffycore_utils.PrettyJSON(jwtJson))

			jwtJson, err = decodeJwt(ctx, createOperatorResponse.SystemAccount.JWT)
			if err != nil {
				log.Error().Err(err).Msg("failed to decode system account jwt")
				return err
			}
			printer.Info(fluffycore_utils.PrettyJSON(jwtJson))
			err = configurator_shared.WriteNatsServerConfig(ctx, CurrentResolverParams.Name, createOperatorResponse)
			if err != nil {
				log.Error().Err(err).Msg("failed to write nats server config")
				return err
			}
			err = configurator_shared.WriteNatsSystemAccountCreds(ctx, CurrentResolverParams.Name, createOperatorResponse)
			if err != nil {
				log.Error().Err(err).Msg("failed to write nats system account creds")
				return err
			}
			err = configurator_shared.WriteNatsAuthAccountCreds(ctx, CurrentResolverParams.Name, createOperatorResponse)
			if err != nil {
				log.Error().Err(err).Msg("failed to write nats auth account creds")
				return err
			}

			err = configurator_shared.WriteMasterServiceConfig(ctx, CurrentResolverParams.Name, createOperatorResponse)
			if err != nil {
				log.Error().Err(err).Msg("failed to write master service config")
				return err
			}
			return nil

		},
	}

	flagName := "name"
	defaultS := CurrentResolverParams.Name
	command.Flags().StringVar(&CurrentResolverParams.Name, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	parentCmd.AddCommand(command)

}
