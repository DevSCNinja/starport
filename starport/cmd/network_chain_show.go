package starportcmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tendermint/starport/starport/pkg/cosmosutil"
	"github.com/tendermint/starport/starport/pkg/entrywriter"
	"github.com/tendermint/starport/starport/pkg/yaml"
	"github.com/tendermint/starport/starport/services/network"
	"github.com/tendermint/starport/starport/services/network/networkchain"
	"github.com/tendermint/starport/starport/services/network/networktypes"
)

type ShowType string

const (
	chainShowInfo     ShowType = "info"
	chainShowGenesis  ShowType = "genesis"
	chainShowAccounts ShowType = "accounts"
	chainShowPeers    ShowType = "peers"
)

var (
	showTypes = map[ShowType]struct{}{
		chainShowInfo:     {},
		chainShowGenesis:  {},
		chainShowAccounts: {},
		chainShowPeers:    {},
	}

	chainAccSummaryHeader = []string{"Genesis Account", "Coins"}
)

// NewNetworkChainShow creates a new chain show command to show
// a chain on SPN.
func NewNetworkChainShow() *cobra.Command {
	c := &cobra.Command{
		Use:   "show [info|genesis|accounts|peers] [launch-id]",
		Short: "Show details of a chain",
		RunE:  networkChainShowHandler,
		Args:  cobra.ExactArgs(2),
	}

	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetHome())

	return c
}

func networkChainShowHandler(cmd *cobra.Command, args []string) error {
	showType := ShowType(args[0])
	if _, ok := showTypes[showType]; !ok {
		cmd.Usage()
		return fmt.Errorf("invalid arg %s", showType)
	}

	nb, err := newNetworkBuilder(cmd)
	if err != nil {
		return err
	}
	defer nb.Cleanup()

	// parse launch ID.
	launchID, err := network.ParseLaunchID(args[1])
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	chainLaunch, err := n.ChainLaunch(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	content := ""
	switch showType {
	case chainShowGenesis:
		content, err = formatChainGenesis(nb, chainLaunch)
	case chainShowInfo:
		content, err = formatChainInfo(cmd.Context(), chainLaunch)
	case chainShowAccounts:
		content, err = formatChainAccounts(cmd.Context(), n, launchID)
	case chainShowPeers:
		content, err = formatChainPeers(cmd.Context(), n, launchID)
	}
	if err != nil {
		return err
	}

	nb.Spinner.Stop()
	fmt.Print(content)
	return nil
}

func formatChainGenesis(nb NetworkBuilder, chainLaunch networktypes.ChainLaunch) (string, error) {
	c, err := nb.Chain(networkchain.SourceLaunch(chainLaunch))
	if err != nil {
		return "", err
	}

	genesisPath, err := c.GenesisPath()
	if err != nil {
		return "", err
	}
	if _, err = os.Stat(genesisPath); os.IsNotExist(err) {
		return "", fmt.Errorf("chain genesis not initialized: %s", genesisPath)
	}
	genesisFile, err := os.ReadFile(genesisPath)
	if err != nil {
		return "", err
	}
	return string(genesisFile), nil
}

func formatChainInfo(ctx context.Context, chainLaunch networktypes.ChainLaunch) (info string, err error) {
	var genesis []byte
	if chainLaunch.GenesisURL != "" {
		genesis, _, err = cosmosutil.GenesisAndHashFromURL(ctx, chainLaunch.GenesisURL)
		if err != nil {
			return "", err
		}
	}
	chainInfo := struct {
		Chain   networktypes.ChainLaunch `json:"Chain"`
		Genesis []byte                   `json:"Genesis"`
	}{
		Chain:   chainLaunch,
		Genesis: genesis,
	}
	return yaml.Marshal(ctx, chainInfo, "$.Genesis")
}

func formatChainAccounts(ctx context.Context, n network.Network, launchID uint64) (string, error) {
	genesisInformation, err := n.GenesisInformation(ctx, launchID)
	if err != nil {
		return "", err
	}

	genesisAccEntries := make([][]string, 0)
	for _, acc := range genesisInformation.GenesisAccounts {
		genesisAccEntries = append(genesisAccEntries, []string{
			acc.Address,
			acc.Coins,
		})
	}
	result := bytes.NewBufferString("")
	err = entrywriter.MustWrite(result, chainAccSummaryHeader, genesisAccEntries...)
	return result.String(), err
}

func formatChainPeers(ctx context.Context, n network.Network, launchID uint64) (string, error) {
	genesisInformation, err := n.GenesisInformation(ctx, launchID)
	if err != nil {
		return "", err
	}

	peers := make([]string, 0)
	for _, acc := range genesisInformation.GenesisValidators {
		peers = append(peers, acc.Peer)
	}

	return fmt.Sprintf("Persistent Peers: %s\n", strings.Join(peers, ",")), nil
}
