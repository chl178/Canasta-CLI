package remove

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/CanastaWiki/Canasta-CLI-Go/cmd/restart"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/farmsettings"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/mediawiki"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"
)

func NewCmdCreate() *cobra.Command {
	var instance config.Installation
	var wikiName string

	addCmd := &cobra.Command{
		Use:   "remove",
		Short: "remove a new wiki from a Canasta instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			fmt.Printf("Removing wiki '%s' from Canasta instance '%s'...\n", wikiName, instance.Id)
			err = RemoveWiki(wikiName, instance)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Done.")
			return nil
		},
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	addCmd.Flags().StringVarP(&wikiName, "wiki", "w", "", "ID of the wiki")
	addCmd.Flags().StringVarP(&instance.Path, "path", "p", pwd, "Path to the new wiki")
	addCmd.Flags().StringVarP(&instance.Id, "id", "i", "", "Canasta instance ID")
	return addCmd
}

// addWiki accepts the Canasta instance ID, the name, domain and path of the new wiki, and the initial admin info, then creates a new wiki in the instance.
func RemoveWiki(name string, instance config.Installation) error {
	var err error
	//Checking Installation existence
	instance, err = canasta.CheckCanastaId(instance)
	if err != nil {
		return err
	}

	//Migrate to the new version Canasta
	err = canasta.MigrateToNewVersion(instance.Path)
	if err != nil {
		return err
	}

	//Checking Running status
	err = orchestrators.CheckRunningStatus(instance.Path, instance.Id, instance.Orchestrator)
	if err != nil {
		return err
	}

	//Checking Wiki existence
	exists, _, err := farmsettings.CheckWiki(instance.Path, name, "", "")
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("A wiki with the name '%s' does not exist", name)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("This will delete the " + name + " in the Wiki farm and the corresponding database. Continue? [y/N] ")
	text, _ := reader.ReadString('\n')
	text = strings.ToLower(strings.TrimSpace(text))

	if text != "y" {
		fmt.Println("Operation cancelled.")
		return nil
	}

	//Remove the wiki
	err = farmsettings.RemoveWiki(name, instance.Path)
	if err != nil {
		return err
	}

	//Install the corresponding Database
	err = mediawiki.RemoveDatabase(instance.Path, name, instance.Orchestrator)
	if err != nil {
		return err
	}

	//Remove the Localsettings
	err = canasta.RemoveSettings(instance.Path, name)
	if err != nil {
		return err
	}

	//Remove the Images
	err = canasta.RemoveImages(instance.Path, name)
	if err != nil {
		return err
	}

	//Rewrite the Caddyfile
	err = canasta.RewriteCaddy(instance.Path)
	if err != nil {
		return err
	}

	//Stop the Canasta Instance
	err = restart.Restart(instance)
	if err != nil {
		return err
	}

	fmt.Println("Successfully Removed the Wiki '" + name + "from Canasta instance '" + instance.Id + "'...")

	return nil
}
