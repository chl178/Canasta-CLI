package canasta

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/execute"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/farmsettings"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/git"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/logging"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"
)

type CanastaVariables struct {
	Id            string
	AdminPassword string
	AdminName     string
}

// CloneStackRepo accept the orchestrator from the cli and pass the corresponding reopository link
// and clones the repo to a new folder in the specified path
func CloneStackRepo(orchestrator, canastaId string, path *string) error {
	*path += "/" + canastaId
	logging.Print(fmt.Sprintf("Cloning the %s stack repo to %s \n", orchestrator, *path))
	repo := orchestrators.GetRepoLink(orchestrator)
	err := git.Cloneb(repo, *path, "CLI")
	return err
}

// if envPath is passed as argument copies the file located at envPath to the installation directory
// else copies .env.example to .env in the installation directory
func CopyEnv(envPath, path, pwd string) error {
	yamlPath := path + "/config/wikis.yaml"

	if envPath == "" {
		envPath = path + "/.env.example"
	} else {
		envPath = pwd + "/" + envPath
	}
	logging.Print(fmt.Sprintf("Copying %s to %s/.env\n", envPath, path))
	err, output := execute.Run("", "cp", envPath, path+"/.env")
	if err != nil {
		return fmt.Errorf(output)
	}
	_, domainNames, _, err := farmsettings.ReadWikisYaml(yamlPath)
	if err != nil {
		return err
	}
	if err := SaveEnvVariable(path+"/.env", "MW_SITE_SERVER", "https://"+domainNames[0]); err != nil {
		return err
	}
	if err := SaveEnvVariable(path+"/.env", "MW_SITE_FQDN", domainNames[0]); err != nil {
		return err
	}
	return nil
}

func CopyYaml(yamlPath, path string) error {
	logging.Print(fmt.Sprintf("Copying %s to %s/config/wikis.yaml\n", yamlPath, path))
	err, output := execute.Run("", "cp", yamlPath, path+"/config/wikis.yaml")
	if err != nil {
		return fmt.Errorf(output)
	}
	return nil
}

func CopySettings(path string) error {
	yamlPath := path + "/config/wikis.yaml"

	logging.Print(fmt.Sprintf("Copying %s to %s/.env\n", yamlPath, path))
	WikiNames, _, _, err := farmsettings.ReadWikisYaml(yamlPath)
	if err != nil {
		return err
	}
	for i := len(WikiNames) - 1; i >= 0; i-- {
		dirPath := path + fmt.Sprintf("/config/%s", WikiNames[i])

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return err
		}

		// Copy SettingsTemplate.php
		err, output := execute.Run("", "cp", path+"/config/SettingsTemplate.php", dirPath+"/Settings.php")
		if err != nil {
			return fmt.Errorf(output)
		}
	}
	return nil
}

func CopySetting(path, name string) error {
	dirPath := path + fmt.Sprintf("/config/%s", name)

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}

	// Copy SettingsTemplate.php
	err, output := execute.Run("", "cp", path+"/config/SettingsTemplate.php", dirPath+"/Settings.php")
	if err != nil {
		return fmt.Errorf(output)
	}
	return nil
}

func RewriteSettings(path string, WikiNames []string) error {
	for _, name := range WikiNames {
		filePath := path + "/config/" + name + "/LocalSettings.php"

		// Read the original file
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Read file line by line
		scanner := bufio.NewScanner(file)
		var lines []string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "#$wgSitename = ;") {
				line = "$wgSitename = \"" + name + "\";"
			}
			if strings.Contains(line, "#$wgMetaNamespace = ;") {
				line = "$wgMetaNamespace = \"" + name + "\";"
			}
			lines = append(lines, line)
		}

		// Create/Truncate the file for writing
		file, err = os.Create(filePath)
		if err != nil {
			return err
		}

		// Write back to file
		writer := bufio.NewWriter(file)
		for _, line := range lines {
			_, err = fmt.Fprintln(writer, line)
			if err != nil {
				return err
			}
		}
		if err = writer.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// Function to remove duplicates from slice
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func RewriteCaddy(path string) error {
	_, ServerNames, _, err := farmsettings.ReadWikisYaml(path + "/config/wikis.yaml")
	if err != nil {
		return err
	}
	filePath := path + "/config/Caddyfile"

	// Remove duplicates from ServerNames
	ServerNames = removeDuplicates(ServerNames)

	// Generate the new server names line
	var newLine strings.Builder
	for i, name := range ServerNames {
		if i > 0 {
			newLine.WriteString(", ")
		}
		newLine.WriteString(name)
		newLine.WriteString(":{$HTTPS_PORT}")
	}

	// Create/Truncate the file for writing
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	// Write back to file
	writer := bufio.NewWriter(file)

	// Write server names line to file
	_, err = fmt.Fprintln(writer, newLine.String())
	if err != nil {
		return err
	}

	// Write empty line to file
	_, err = fmt.Fprintln(writer, "")
	if err != nil {
		return err
	}

	// Write reverse proxy line to file
	_, err = fmt.Fprintln(writer, "reverse_proxy varnish:80")
	if err != nil {
		return err
	}

	if err = writer.Flush(); err != nil {
		return err
	}

	return nil
}

// Copies the LocalSettings.php at localSettingsPath to /config at the installation directory
func CopyLocalSettings(localSettingsPath, path, pwd string) error {
	if localSettingsPath != "" {
		localSettingsPath = pwd + "/" + localSettingsPath
		logging.Print(fmt.Sprintf("Copying %s to %s/config/LocalSettings.php\n", localSettingsPath, path))
		err, output := execute.Run("", "cp", localSettingsPath, path+"/config/LocalSettings.php")
		if err != nil {
			logging.Fatal(fmt.Errorf(output))
		}
	}
	return nil
}

// Copies database dump from databasePath to the /_initdb/ at the installation directory
func CopyDatabase(databasePath, path, pwd string) error {
	if databasePath != "" {
		databasePath = pwd + "/" + databasePath
		logging.Print(fmt.Sprintf("Copying %s to %s/_initdb\n", databasePath, path))
		err, output := execute.Run("", "cp", databasePath, path+"/_initdb/")
		if err != nil {
			logging.Fatal(fmt.Errorf(output))
		}
	}
	return nil
}

// Verifying file extension for database dump
func SanityChecks(databasePath, localSettingsPath string) error {
	if databasePath == "" {
		return fmt.Errorf("database dump path not mentioned")
	}
	if localSettingsPath == "" {
		return fmt.Errorf("localsettings.php path not mentioned")
	}
	if !strings.HasSuffix(databasePath, ".sql") && !strings.HasSuffix(databasePath, ".sql.gz") {
		return fmt.Errorf("mysqldump is of invalid file type")
	}
	if !strings.HasSuffix(localSettingsPath, ".php") {
		return fmt.Errorf("make sure correct LocalSettings.php is mentioned")
	}
	return nil
}

// Make changes to the .env file at the installation directory
func SaveEnvVariable(envPath, key, value string) error {
	file, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}
	data := string(file)
	list := strings.Split(data, "\n")
	for index, line := range list {
		if strings.Contains(line, key) {
			list[index] = fmt.Sprintf("%s=%s", key, value)
		}
	}
	lines := strings.Join(list, "\n")
	if err := ioutil.WriteFile(envPath, []byte(lines), 0644); err != nil {
		return err
	}
	return nil
}

// Get values saved inside the .env at the envPath
func GetEnvVariable(envPath string) map[string]string {
	EnvVariables := make(map[string]string)
	file_data, err := os.ReadFile(envPath)
	if err != nil {
		logging.Fatal(err)
	}
	data := strings.TrimSuffix(string(file_data), "\n")
	variable_list := strings.Split(data, "\n")
	for _, variable := range variable_list {
		list := strings.Split(variable, "=")
		if len(list) < 2 {
			continue
		}
		EnvVariables[list[0]] = list[1]
	}
	return EnvVariables
}

// Checking Installation existence
func CheckCanastaId(instance config.Installation) (config.Installation, error) {
	var err error
	if instance.Id != "" {
		if instance, err = config.GetDetails(instance.Id); err != nil {
			return instance, err
		}
	} else {
		if instance.Id, err = config.GetCanastaId(instance.Path); err != nil {
			return instance, err
		}
		if instance, err = config.GetDetails(instance.Id); err != nil {
			return instance, err
		}
	}
	return instance, nil
}

func DeleteConfigAndContainers(keepConfig bool, installationDir, orchestrator string) {
	fmt.Println("Removing containers")
	orchestrators.DeleteContainers(installationDir, orchestrator)
	fmt.Println("Deleting config files")
	orchestrators.DeleteConfig(installationDir)
	fmt.Println("Deleted all containers and config files")
}

func RemoveSettings(path, name string) error {
	// Prepare the file path
	filePath := filepath.Join(path, "config", name)

	// Check if the file exists
	if _, err := os.Stat(filePath); err == nil {
		// If the file exists, remove it
		err, output := execute.Run("", "rm", "-rf", filePath)
		if err != nil {
			return fmt.Errorf(output)
		}
	} else if os.IsNotExist(err) {
		// File does not exist, do nothing
		return nil
	} else {
		// File may or may not exist. See the specific error
		return err
	}
	return nil
}

func RemoveImages(path, name string) error {
	// Prepare the file path
	filePath := filepath.Join(path, "images", name)

	// Check if the file exists
	if _, err := os.Stat(filePath); err == nil {
		// If the file exists, remove it
		err, output := execute.Run("", "rm", "-rf", filePath)
		if err != nil {
			return fmt.Errorf(output)
		}
	} else if os.IsNotExist(err) {
		// File does not exist, do nothing
		return nil
	} else {
		// File may or may not exist. See the specific error
		return err
	}
	return nil
}

func MigrateToNewVersion(path string) error {
	// Determine the path to the wikis.yaml file
	yamlPath := filepath.Join(path, "config", "wikis.yaml")

	// Check if the file already exists
	if _, err := os.Stat(yamlPath); err == nil {
		// File exists, assume the user is already using the new configuration
		return nil
	} else if !os.IsNotExist(err) {
		// If the error is not because the file does not exist, return it
		return err
	}

	// Open the .env file
	envFile, err := os.Open(filepath.Join(path, ".env"))
	if err != nil {
		return err
	}
	defer envFile.Close()

	// Read the environment variables from the .env file
	envMap := make(map[string]string)
	scanner := bufio.NewScanner(envFile)
	name := "my_wiki"

	for scanner.Scan() {
		line := scanner.Text()
		splitLine := strings.SplitN(line, "=", 2)
		if len(splitLine) != 2 {
			continue
		}
		envMap[splitLine[0]] = splitLine[1]
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Remove the "http://" or "https://" prefix from MW_SITE_SERVER variable
	mwSiteServer := strings.TrimPrefix(envMap["MW_SITE_SERVER"], "http://")
	mwSiteServer = strings.TrimPrefix(mwSiteServer, "https://")

	// Create the wikis.yaml file using farmsettings.GenerateWikisYaml
	_, err = farmsettings.GenerateWikisYaml(yamlPath, name, mwSiteServer)
	if err != nil {
		return err
	}

	return nil
}
