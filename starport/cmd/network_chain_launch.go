package starportcmd

import (
	"github.com/spf13/cobra"
	"github.com/tendermint/starport/starport/services/network"
)

const (
	flagRemainingTime = "remaining-time"
)

// NewNetworkChainLaunch creates a new chain launch command to launch
// the network as a coordinator.
func NewNetworkChainLaunch() *cobra.Command {
	c := &cobra.Command{
		Use:   "launch [launch-id]",
		Short: "Launch a network as a coordinator",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainLaunchHandler,
	}

	c.Flags().Duration(flagRemainingTime, 0, "The remaining time for validator preparation before the chain is effectively launched")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetKeyringBackend())

	return c
}

func networkChainLaunchHandler(cmd *cobra.Command, args []string) error {
	nb, err := newNetworkBuilder(cmd)
	if err != nil {
		return err
	}
	defer nb.Cleanup()

	// parse launch ID
	launchID, err := network.ParseLaunchID(args[0])
	if err != nil {
		return err
	}

	remainingTime, _ := cmd.Flags().GetDuration(flagRemainingTime)

	n, err := nb.Network()
	if err != nil {
		return err
	}

	return n.TriggerLaunch(cmd.Context(), launchID, remainingTime)
}
