package extension

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
	instance     config.Installation
	pwd          string
	wiki		 string
	err          error
	verbose      bool
	extensionCmd *cobra.Command
	constants    = extensionsskins.Item{Name: "Canasta extension", RelativeInstallationPath: "canasta-extensions", PhpCommand: "wfLoadExtension"}
)

func NewCmdCreate() *cobra.Command {
	extensionCmd = &cobra.Command{
		Use:   "extension",
		Short: "Manage Canasta extensions",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logging.SetVerbose(verbose)
			instance, err = canasta.CheckCanastaId(instance)
			if err != nil {
				return err
			}
			return nil
		},
	}

	if pwd, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}
	extensionCmd.PersistentFlags().StringVarP(&instance.Id, "id", "i", "", "Canasta instance ID")
	extensionCmd.PersistentFlags().StringVarP(&instance.Path, "path", "p", pwd, "Canasta installation directory")
	extensionCmd.PersistentFlags().StringVarP(&wiki, "wiki", "w", "", "ID of the specific wiki within the Canasta farm")
	extensionCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose Output")

	extensionCmd.AddCommand(listCmdCreate())
	extensionCmd.AddCommand(enableCmdCreate())
	extensionCmd.AddCommand(disableCmdCreate())

	return extensionCmd
}
