package cli

import (
	"github.com/spf13/cobra"
	"github.com/ljjjustin/themis/client"
	"os"
	"fmt"
)

var configFile string

func ExecuateFencerCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "executeFencer <host id>",
		Short: "execute fence a host",
		Run:   executeFencerFunc,
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to toml config file.")

	return &cmd
}

func executeFencerFunc(cmd *cobra.Command, args []string) {

	themis := client.NewThemisClient(globalFlags.Url)

	err := themis.ExecuteFencerFunc(getHostId(args), configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	} else {
		fmt.Println("fence success.")
	}
}