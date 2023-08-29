package skin

import (
	"fmt"
	"strings"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/extensionsskins"
	"github.com/spf13/cobra"
)

func enableCmdCreate() *cobra.Command {

	enableCmd := &cobra.Command{
		Use:   "enable SKIN_NAME",
		Short: "Enable a Canasta skin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skins := strings.Split(args[0], ",")
			for _, skin := range skins {
				skinName, err := extensionsskins.CheckInstalled(skin, instance, constants)
				if err != nil {
					fmt.Print(err.Error() + "\n")
					continue
				}
				extensionsskins.Enable(skinName, wiki, instance, constants)
			}
			return err
		},
	}
	return enableCmd
}
