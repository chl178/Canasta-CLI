package skin

import (
	"fmt"
	"strings"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/extensionsskins"
	"github.com/spf13/cobra"
)

func disableCmdCreate() *cobra.Command {

	disableCmd := &cobra.Command{
		Use:   "disable SKIN_NAME",
		Short: "Disable a Canasta skin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skins := strings.Split(args[0], ",")
			for _, skin := range skins {
				skinName, err := extensionsskins.CheckEnabled(skin, wiki, instance, constants)
				if err != nil {
					fmt.Print(err.Error() + "\n")
					continue
				}
				extensionsskins.Disable(skinName, wiki, instance, constants)
			}
			return err
		},
	}
	return disableCmd
}
