package farmsettings

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Wiki struct {
	ID   string `yaml:"id"`
	URL  string `yaml:"url"`
	NAME string `yaml:"name"`
}

type Wikis struct {
	Wikis []Wiki `yaml:"wikis"`
}

func CreateYaml(name, domain string, path *string) error {
	if *path == "" {
		var err error
		*path, err = GenerateWikisYaml("./wikis.yaml", name, domain)
		if err != nil {
			return err
		}
	}

	return nil
}

func GenerateWikisYaml(filePath, name, domain string) (string, error) {
	wikis := Wikis{}
	wikis.Wikis = append(wikis.Wikis, Wiki{ID: name, URL: domain, NAME: name})

	out, err := yaml.Marshal(&wikis)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filePath, out, 0644)
	if err != nil {
		return "", err
	}

	fmt.Println("Successfully written to wikis.yaml")
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

func ReadWikisYaml(filePath string) ([]string, []string, []string, error) {
	// Read the YAML file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, nil, err
	}

	// Parse the YAML data into a Wikis struct
	wikis := Wikis{}
	err = yaml.Unmarshal(data, &wikis)
	if err != nil {
		return nil, nil, nil, err
	}

	// Check if any wikis exist
	if len(wikis.Wikis) == 0 {
		return nil, nil, nil, fmt.Errorf("no wikis found in the YAML file")
	}

	// Extract IDs, server names, and paths from URLs
	ids := make([]string, 0, len(wikis.Wikis))
	serverNames := make([]string, 0, len(wikis.Wikis))
	paths := make([]string, 0, len(wikis.Wikis))
	for _, wiki := range wikis.Wikis {
		ids = append(ids, wiki.ID)

		// Split the URL at the first '/' to extract the server name and path
		parts := strings.SplitN(wiki.URL, "/", 2)
		serverName := parts[0]
		serverNames = append(serverNames, serverName)
		if len(parts) > 1 {
			paths = append(paths, "/"+parts[1])
		} else {
			paths = append(paths, "/")
		}
	}

	return ids, serverNames, paths, nil
}

func CheckWiki(path, name, domain, wikiPath string) (bool, bool, error) {
	// Get the absolute path to the wikis.yaml file
	filePath := filepath.Join(path, "config", "wikis.yaml")

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File does not exist
		return false, false, nil
	}

	// Read the wikis from the YAML file
	ids, serverNames, paths, err := ReadWikisYaml(filePath)
	if err != nil {
		return false, false, err
	}

	// Variables to hold whether the wiki name and path are found
	nameExists := false
	pathComboExists := false

	// Check if a wiki with the given name exists
	for i, id := range ids {
		if id == name {
			nameExists = true
		}
		if serverNames[i]+paths[i] == domain+"/"+wikiPath {
			pathComboExists = true
		}
	}

	return nameExists, pathComboExists, nil
}

func AddWiki(name, path, domain, wikipath, siteName string) error {
	// Get the absolute path to the wikis.yaml file
	filePath := filepath.Join(path, "config", "wikis.yaml")

	if siteName == "" {
		siteName = name
	}
	// Read the existing wikis from the YAML file
	wikis := Wikis{}

	// Check if the file exists before trying to read it
	if _, err := os.Stat(filePath); err == nil {
		// File exists, read it
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Unmarshal the yaml file content into wikis
		err = yaml.Unmarshal(data, &wikis)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		// If the error is not because the file does not exist, return it
		return err
	}

	// Create a new wiki
	newWiki := Wiki{ID: name, URL: filepath.Join(domain, wikipath), NAME: siteName}

	// Append the new wiki to the list of wikis
	wikis.Wikis = append(wikis.Wikis, newWiki)

	// Marshal the wikis back into YAML
	updatedData, err := yaml.Marshal(wikis)
	if err != nil {
		return err
	}

	// Write the updated data back to the file
	err = ioutil.WriteFile(filePath, updatedData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func RemoveWiki(name, path string) error {
	// Get the absolute path to the wikis.yaml file
	filePath := filepath.Join(path, "config", "wikis.yaml")

	// Read the existing wikis from the YAML file
	wikis := Wikis{}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal the yaml file content into wikis
	err = yaml.Unmarshal(data, &wikis)
	if err != nil {
		return err
	}

	// Initialize an empty slice to store the remaining wikis
	remainingWikis := []Wiki{}

	// Find and remove the specified wiki
	for _, wiki := range wikis.Wikis {
		if wiki.ID != name {
			remainingWikis = append(remainingWikis, wiki)
		}
	}

	// Check if all wikis were removed
	if len(remainingWikis) == 0 {
		return fmt.Errorf("cannot remove the last wiki in the Canasta Instance")
	}

	// Replace the existing wikis with the remaining wikis
	wikis.Wikis = remainingWikis

	// Marshal the updated wikis back into YAML
	updatedData, err := yaml.Marshal(wikis)
	if err != nil {
		return err
	}

	// Write the updated data back to the file
	err = ioutil.WriteFile(filePath, updatedData, 0644)
	if err != nil {
		return err
	}

	return nil
}
