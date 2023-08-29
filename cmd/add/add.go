package add

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/CanastaWiki/Canasta-CLI-Go/cmd/restart"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/farmsettings"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/mediawiki"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/prompt"
)

func NewCmdCreate() *cobra.Command {
	var instance config.Installation
	var wikiName string
	var domainName string
	var wikiPath string
	var siteName string
	var databasePath string
	var url string

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new wiki to a Canasta instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			wikiName, domainName, wikiPath, instance.Id, siteName, err = prompt.PromptWiki(wikiName, url, instance.Id, siteName)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Adding wiki '%s' to Canasta instance '%s'...\n", wikiName, instance.Id)
			err = AddWiki(wikiName, domainName, wikiPath, siteName, databasePath, instance)
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

	addCmd.Flags().StringVarP(&wikiName, "wiki", "w", "", "ID of the new wiki")
	addCmd.Flags().StringVarP(&url, "url", "u", "", "URL of the new wiki")
	addCmd.Flags().StringVarP(&siteName, "site-name", "s", "", "Name of the new wiki site")
	addCmd.Flags().StringVarP(&instance.Path, "path", "p", pwd, "Path to the new wiki")
	addCmd.Flags().StringVarP(&instance.Id, "id", "i", "", "Canasta instance ID")
	addCmd.Flags().StringVarP(&instance.Orchestrator, "orchestrator", "o", "docker-compose", "Orchestrator to use for installation")
	addCmd.Flags().StringVarP(&databasePath, "database", "d", "", "Path to the existing database dump")
	return addCmd
}

// addWiki accepts the Canasta instance ID, the name, domain and path of the new wiki, and the initial admin info, then creates a new wiki in the instance.
func AddWiki(name, domain, wikipath, siteName, databasePath string, instance config.Installation) error {
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
	exists, pathComboExists, err := farmsettings.CheckWiki(instance.Path, name, domain, wikipath)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("A wiki with the name '%s' exists", name)
	}
	if pathComboExists {
		return fmt.Errorf("A wiki with the same installation path '%s' in the Canasta '%s' exists", name+": "+domain+"/"+wikipath, instance.Id)
	}

	//Add the wiki in farmsettings
	err = farmsettings.AddWiki(name, instance.Path, domain, wikipath, siteName)
	if err != nil {
		return err
	}

	// Import the database if databasePath is specified
	if databasePath != "" {
		err = orchestrators.ImportDatabase(name, databasePath, instance)
		if err != nil {
			return err
		}
	}

	//Copy the Localsettings
	err = canasta.CopySetting(instance.Path, name)
	if err != nil {
		return err
	}

	//Rewrite the Caddyfile
	err = canasta.RewriteCaddy(instance.Path)
	if err != nil {
		return err
	}

	err = mediawiki.InstallOne(instance.Path, name, domain, wikipath, instance.Orchestrator)
	if err != nil {
		return err
	}
	err = restart.Restart(instance)
	if err != nil {
		return err
	}

	fmt.Println("Successfully Added the Wiki '" + name + "in Canasta instance '" + instance.Id + "'...")

	return nil
}
