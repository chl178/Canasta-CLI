package extension

import (
	"fmt"
	"strings"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/extensionsskins"
	"github.com/spf13/cobra"
)

func disableCmdCreate() *cobra.Command {

	disableCmd := &cobra.Command{
		Use:   "disable EXTENSION1,EXTENSION2,...",
		Short: "Disable a Canasta extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions := strings.Split(args[0], ",")
			for _, extension := range extensions {
				extensionName, err := extensionsskins.CheckEnabled(extension, wiki, instance, constants)
				if err != nil {
					fmt.Print(err.Error() + "\n")
					continue
				}
				extensionsskins.Disable(extensionName, wiki, instance, constants)
			}
			return err
		},
	}
	return disableCmd
}
