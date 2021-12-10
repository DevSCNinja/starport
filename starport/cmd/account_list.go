package starportcmd

import (
	"github.com/spf13/cobra"
	"github.com/tendermint/starport/starport/pkg/cosmosaccount"
)

func NewAccountList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "Show a list of all accounts",
		RunE:  accountListHandler,
	}

	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetAccountPrefixes())

	return c
}

func accountListHandler(cmd *cobra.Command, args []string) error {
	ca, err := cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(getKeyringBackend(cmd)),
	)
	if err != nil {
		return err
	}

	accounts, err := ca.List()
	if err != nil {
		return err
	}

	return printAccounts(cmd, accounts...)
}
