package url_resolver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	cobra_utils "natsauth/internal/cobra_utils"
	"net/http"
	"os"
	"strings"
	"sync"

	shared "natsauth/internal/shared"

	callout_shared "natsauth/cmd/cli/root/callout/shared"

	status "github.com/gogo/status"
	echo "github.com/labstack/echo/v4"
	nkeys "github.com/nats-io/nkeys"
	zerolog "github.com/rs/zerolog"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
	codes "google.golang.org/grpc/codes"
)

const use = "url_resolver"

var (
	appInputs    = shared.NewInputs()
	accountMutex = sync.Mutex{}
	port         = 4299
)

// will be in a persistent store
type AccountManager struct {
	accountMutex                     sync.Mutex
	IssuerKeyPair                    nkeys.KeyPair
	accountFriendlyNameToAccountInfo map[string]*callout_shared.CreateSimpleAccountResponse
	accountPubKeyToAccountInfo       map[string]*callout_shared.CreateSimpleAccountResponse
}

func NewAccountManager(IssuerKeyPair nkeys.KeyPair) *AccountManager {
	am := &AccountManager{
		IssuerKeyPair:                    IssuerKeyPair,
		accountFriendlyNameToAccountInfo: make(map[string]*callout_shared.CreateSimpleAccountResponse),
		accountPubKeyToAccountInfo:       make(map[string]*callout_shared.CreateSimpleAccountResponse),
	}

	return am
}
func (s *AccountManager) GetAccounts() map[string]*callout_shared.CreateSimpleAccountResponse {
	//--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
	s.accountMutex.Lock()
	defer s.accountMutex.Unlock()
	//--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
	return s.accountFriendlyNameToAccountInfo
}
func (s *AccountManager) AddAuthAccount(ctx context.Context, jwt string) error {
	log := zerolog.Ctx(ctx).With().Str("func", "AddAuthAccount").Logger()
	//--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
	s.accountMutex.Lock()
	defer s.accountMutex.Unlock()
	//--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
	subI, err := shared.ExtractClaimFromJWTNoValidation(jwt, "sub")
	if err != nil {
		log.Error().Err(err).Msg("error extracting sub from jwt")
		return err
	}
	authAudience := subI.(string)
	s.accountFriendlyNameToAccountInfo["auth"] = &callout_shared.CreateSimpleAccountResponse{
		CommonAccountData: callout_shared.CommonAccountData{
			Name:     "auth",
			JWT:      string(jwt),
			Audience: authAudience,
		},
	}
	s.accountPubKeyToAccountInfo[authAudience] = s.accountFriendlyNameToAccountInfo["AUTH"]
	return nil
}
func (s *AccountManager) GetAccountById(ctx context.Context, id string) (*callout_shared.CreateSimpleAccountResponse, error) {
	//--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
	s.accountMutex.Lock()
	defer s.accountMutex.Unlock()
	//--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
	createSimpleAccountResponse, ok := s.accountPubKeyToAccountInfo[id]
	if ok {
		return createSimpleAccountResponse, nil
	}
	return nil, status.Error(codes.NotFound, "account not found")
}
func (s *AccountManager) GetOrCreateAccountByFriendlyName(ctx context.Context, name string) (*callout_shared.CreateSimpleAccountResponse, error) {
	//--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
	s.accountMutex.Lock()
	defer s.accountMutex.Unlock()
	//--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
	createSimpleAccountResponse, ok := s.accountFriendlyNameToAccountInfo[name]
	if ok {
		return createSimpleAccountResponse, nil
	}
	var err error
	createSimpleAccountResponse, err = callout_shared.CreateSimpleAccount(ctx,
		&callout_shared.CreateSimpleAccountRequest{
			Name:          name,
			IssuerKeyPair: s.IssuerKeyPair,
		})
	if err != nil {
		return nil, err
	}
	s.accountFriendlyNameToAccountInfo[name] = createSimpleAccountResponse
	s.accountPubKeyToAccountInfo[createSimpleAccountResponse.Audience] = createSimpleAccountResponse

	return createSimpleAccountResponse, err
}

var wellknownAccounts = []string{"svc", "edge"}

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

			authAccountJWT, err := os.ReadFile(appInputs.AuthAccountJWTFile)
			if err != nil {
				return err
			}

			okp, err := loadAndParseKeys(appInputs.OperatorNKeyFile, 'O')
			if err != nil {
				log.Error().Err(err).Msg("error loading operator key")
				return err
			}
			accountManager := NewAccountManager(okp)
			accountManager.AddAuthAccount(ctx, string(authAccountJWT))

			for _, wa := range wellknownAccounts {
				accountManager.GetOrCreateAccountByFriendlyName(ctx, wa)
			}

			e := echo.New()

			// Route to get account by ID
			e.GET("/jwt/v1/accounts/id/:id", func(c echo.Context) error {
				//--~--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
				accountMutex.Lock()
				defer accountMutex.Unlock()
				//--~--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
				id := c.Param("id")

				accountInfo, err := accountManager.GetAccountById(ctx, id)
				if err != nil {
					return c.String(http.StatusInternalServerError, "Error creating account: "+err.Error())
				}
				theJWT := accountInfo.JWT
				if theJWT == "" {
					return c.String(http.StatusNotFound, "Account JWT not found")
				}

				return c.String(http.StatusOK, theJWT)
			})

			// Route to get account by name
			e.GET("/jwt/v1/accounts/name/:name", func(c echo.Context) error {

				//--~--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
				accountMutex.Lock()
				defer accountMutex.Unlock()
				//--~--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
				name := c.Param("name")
				name = strings.ToLower(name)
				info, err := accountManager.GetOrCreateAccountByFriendlyName(ctx, name)
				if err != nil {
					return c.String(http.StatusInternalServerError, "Error creating account: "+err.Error())
				}

				id := info.Audience
				if id == "" {
					return c.String(http.StatusNotFound, "Account ID not found")
				}
				// return just the id
				return c.String(http.StatusOK, id)
			})

			// Route to get all accounts
			e.GET("/jwt/v1/accounts", func(c echo.Context) error {
				return c.JSON(http.StatusOK, accountManager.GetAccounts())
			})

			address := fmt.Sprintf(":%d", port)
			printer.Infof("Starting server on %s", address)
			// Start the server on port 8080
			e.Start(address)
			return nil
		},
	}
	shared.InitCommonConnFlags(appInputs, command)

	flagName := "operator.nk"
	defaultS := "operator.nk"
	command.Flags().StringVar(&appInputs.OperatorNKeyFile, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "port"
	defaultInt := port
	command.Flags().IntVar(&defaultInt, flagName, defaultInt, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "auth.account.jwt"
	defaultS = "auth.account.jwt"
	command.Flags().StringVar(&appInputs.AuthAccountJWTFile, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	parentCmd.AddCommand(command)

}

func loadAndParseKeys(fp string, kind byte) (nkeys.KeyPair, error) {
	if fp == "" {
		return nil, errors.New("key file required")
	}
	seed, err := os.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("error reading key file: %w", err)
	}
	if !bytes.HasPrefix(seed, []byte{'S', kind}) {
		return nil, fmt.Errorf("key must be a private key")
	}
	kp, err := nkeys.FromSeed(seed)
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %w", err)
	}
	return kp, nil
}
