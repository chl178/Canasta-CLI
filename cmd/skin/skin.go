package skin

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/extensionsskins"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/logging"
)

var (
	instance  config.Installation
	pwd       string
	wiki      string
	err       error
	verbose   bool
	skinCmd   *cobra.Command
	constants = extensionsskins.Item{Name: "Canasta skin", RelativeInstallationPath: "canasta-skins", PhpCommand: "wfLoadSkin"}
)

func NewCmdCreate() *cobra.Command {
	skinCmd = &cobra.Command{
		Use:   "skin",
		Short: "Manage Canasta skins",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logging.SetVerbose(verbose)
			instance, err = canasta.CheckCanastaId(instance)
			return err
		},
	}

	if pwd, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}
	skinCmd.PersistentFlags().StringVarP(&instance.Id, "id", "i", "", "Canasta instance ID")
	skinCmd.PersistentFlags().StringVarP(&instance.Path, "path", "p", pwd, "Canasta installation directory")
	skinCmd.PersistentFlags().StringVarP(&wiki, "wiki", "w", "", "ID of the specific wiki within the Canasta farm")
	skinCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose Output")

	skinCmd.AddCommand(listCmdCreate())
	skinCmd.AddCommand(enableCmdCreate())
	skinCmd.AddCommand(disableCmdCreate())

	return skinCmd
}
