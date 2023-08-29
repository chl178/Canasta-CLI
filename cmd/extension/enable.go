package extension

import (
	"fmt"
	"strings"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/extensionsskins"
	"github.com/spf13/cobra"
)

func enableCmdCreate() *cobra.Command {

	enableCmd := &cobra.Command{
		Use:   "enable EXTENSION1,EXTENSION2,...",
		Short: "Enable a Canasta extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions := strings.Split(args[0], ",")
			for _, extension := range extensions {
				extensionName, err := extensionsskins.CheckInstalled(extension, instance, constants)
				if err != nil {
					fmt.Print(err.Error() + "\n")
					continue
				}
				extensionsskins.Enable(extensionName, wiki, instance, constants)
			}
			return err
		},
	}
	return enableCmd
}
