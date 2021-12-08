package starportcmd

import (
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/spf13/cobra"
	"github.com/tendermint/starport/starport/services/chain"
)

const (
	flagSimappGenesis                = "genesis"
	flagSimappParams                 = "params"
	flagSimappExportParamsPath       = "exportParamsPath"
	flagSimappExportParamsHeight     = "exportParamsHeight"
	flagSimappExportStatePath        = "exportStatePath"
	flagSimappExportStatsPath        = "exportStatsPath"
	flagSimappSeed                   = "seed"
	flagSimappInitialBlockHeight     = "initialBlockHeight"
	flagSimappNumBlocks              = "numBlocks"
	flagSimappBlockSize              = "blockSize"
	flagSimappLean                   = "lean"
	flagSimappCommit                 = "commit"
	flagSimappSimulateEveryOperation = "simulateEveryOperation"
	flagSimappPrintAllInvariants     = "printAllInvariants"
	flagSimappVerbose                = "verbose"
	flagSimappPeriod                 = "period"
	flagSimappGenesisTime            = "genesisTime"
)

// NewChainSimulate creates a new simulation command to run the blockchain simulation.
func NewChainSimulate() *cobra.Command {
	c := &cobra.Command{
		Use:   "simulate",
		Short: "Run the blockchain simulation node in development",
		Long:  "Run the blockchain simulation for all chain modules",
		Args:  cobra.ExactArgs(0),
		RunE:  chainSimulationHandler,
	}
	simappFlags(c)
	flagSetPath(c)
	c.Flags().AddFlagSet(flagSetHome())
	return c
}

func chainSimulationHandler(cmd *cobra.Command, args []string) error {
	var (
		// simulation flags
		verbose, _     = cmd.Flags().GetBool(flagSimappVerbose)
		period, _      = cmd.Flags().GetUint(flagSimappPeriod)
		genesisTime, _ = cmd.Flags().GetInt64(flagSimappGenesisTime)
		config         = newConfigFromFlags(cmd)
	)

	// create the chain
	c, err := newChainWithHomeFlags(cmd)
	if err != nil {
		return err
	}

	return c.Simulate(cmd.Context(),
		chain.SimappWithVerbose(verbose),
		chain.SimappWithPeriod(period),
		chain.SimappWithGenesisTime(genesisTime),
		chain.SimappWithConfig(config),
	)
}

// newConfigFromFlags creates a simulation from the retrieved values of the flags.
func newConfigFromFlags(cmd *cobra.Command) simulation.Config {
	var (
		genesis, _                = cmd.Flags().GetString(flagSimappGenesis)
		params, _                 = cmd.Flags().GetString(flagSimappParams)
		exportParamsPath, _       = cmd.Flags().GetString(flagSimappExportParamsPath)
		exportParamsHeight, _     = cmd.Flags().GetInt(flagSimappExportParamsHeight)
		exportStatePath, _        = cmd.Flags().GetString(flagSimappExportStatePath)
		exportStatsPath, _        = cmd.Flags().GetString(flagSimappExportStatsPath)
		seed, _                   = cmd.Flags().GetInt64(flagSimappSeed)
		initialBlockHeight, _     = cmd.Flags().GetInt(flagSimappInitialBlockHeight)
		numBlocks, _              = cmd.Flags().GetInt(flagSimappNumBlocks)
		blockSize, _              = cmd.Flags().GetInt(flagSimappBlockSize)
		lean, _                   = cmd.Flags().GetBool(flagSimappLean)
		commit, _                 = cmd.Flags().GetBool(flagSimappCommit)
		simulateEveryOperation, _ = cmd.Flags().GetBool(flagSimappSimulateEveryOperation)
		printAllInvariants, _     = cmd.Flags().GetBool(flagSimappPrintAllInvariants)
	)
	return simulation.Config{
		GenesisFile:        genesis,
		ParamsFile:         params,
		ExportParamsPath:   exportParamsPath,
		ExportParamsHeight: exportParamsHeight,
		ExportStatePath:    exportStatePath,
		ExportStatsPath:    exportStatsPath,
		Seed:               seed,
		InitialBlockHeight: initialBlockHeight,
		NumBlocks:          numBlocks,
		BlockSize:          blockSize,
		Lean:               lean,
		Commit:             commit,
		OnOperation:        simulateEveryOperation,
		AllInvariants:      printAllInvariants,
	}
}

func simappFlags(c *cobra.Command) {
	// config fields
	c.Flags().String(flagSimappGenesis, "", "custom simulation genesis file; cannot be used with params file")
	c.Flags().String(flagSimappParams, "", "custom simulation params file which overrides any random params; cannot be used with genesis")
	c.Flags().String(flagSimappExportParamsPath, "", "custom file path to save the exported params JSON")
	c.Flags().Int(flagSimappExportParamsHeight, 0, "height to which export the randomly generated params")
	c.Flags().String(flagSimappExportStatePath, "", "custom file path to save the exported app state JSON")
	c.Flags().String(flagSimappExportStatsPath, "", "custom file path to save the exported simulation statistics JSON")
	c.Flags().Int64(flagSimappSeed, 42, "simulation random seed")
	c.Flags().Int(flagSimappInitialBlockHeight, 1, "initial block to start the simulation")
	c.Flags().Int(flagSimappNumBlocks, 200, "number of new blocks to simulate from the initial block height")
	c.Flags().Int(flagSimappBlockSize, 30, "operations per block")
	c.Flags().Bool(flagSimappLean, false, "lean simulation log output")
	c.Flags().Bool(flagSimappCommit, true, "have the simulation commit")
	c.Flags().Bool(flagSimappSimulateEveryOperation, false, "run slow invariants every operation")
	c.Flags().Bool(flagSimappPrintAllInvariants, false, "print all invariants if a broken invariant is found")

	// simulation flags
	c.Flags().BoolP(flagSimappVerbose, "v", false, "verbose log output")
	c.Flags().Uint(flagSimappPeriod, 0, "run slow invariants only once every period assertions")
	c.Flags().Int64(flagSimappGenesisTime, 0, "override genesis UNIX time instead of using a random UNIX time")
}
