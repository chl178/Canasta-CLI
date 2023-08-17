package mediawiki

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/execute"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/farmsettings"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/logging"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"
	"github.com/sethvargo/go-password/password"
)

const dbServer = "db"
const confPath = "/mediawiki/config/"
const scriptPath = "/w"

func Install(path, yamlPath, orchestrator string, canastaInfo canasta.CanastaVariables) (canasta.CanastaVariables, error) {
	var err error
	logging.Print("Configuring MediaWiki Installation\n")
	logging.Print("Running install.php\n")
	envVariables := canasta.GetEnvVariable(path + "/.env")
	settingsName := "CommonSettings.php"

	command := "/wait-for-it.sh -t 60 db:3306"
	output, err := orchestrators.ExecWithError(path, orchestrator, "web", command)
	if err != nil {
		return canastaInfo, fmt.Errorf(output)
	}
	if canastaInfo.AdminPassword == "" {
		canastaInfo.AdminPassword, err = password.Generate(12, 2, 4, false, true)
		if err != nil {
			return canastaInfo, err
		}
		// Save automatically generated password to .admin-password inside the configuration folder
		fmt.Printf("Saving password to %s/.admin-password\n", path)
		file, err := os.Create(path + "/.admin-password")
		if err != nil {
			return canastaInfo, err
		}
		defer file.Close()
		_, err = file.WriteString(canastaInfo.AdminPassword)
		if err != nil {
			return canastaInfo, err
		}
	}

	fmt.Printf("Saving adminname to %s/.admin\n", path)
	file, err := os.Create(path + "/.admin")
	if err != nil {
		return canastaInfo, err
	}
	defer file.Close()
	_, err = file.WriteString(canastaInfo.AdminName)
	if err != nil {
		return canastaInfo, err
	}

	WikiNames, domainNames, _, err := farmsettings.ReadWikisYaml(yamlPath)
	if err != nil {
		return canastaInfo, err
	}
	for i := 0; i < len(WikiNames); i++ {
		wikiName := WikiNames[i]
		domainName := domainNames[i]

		command := fmt.Sprintf("php maintenance/install.php --dbserver=%s --dbname='%s' --confpath=%s --scriptpath=%s --server='https://%s' --dbuser='%s' --dbpass='%s'  --pass='%s' '%s' '%s'",
			dbServer, wikiName, confPath, scriptPath, domainName, "root", envVariables["MYSQL_PASSWORD"], canastaInfo.AdminPassword, wikiName, canastaInfo.AdminName)

		output, err = orchestrators.ExecWithError(path, orchestrator, "web", command)
		if err != nil {
			return canastaInfo, fmt.Errorf(output)
		}
		time.Sleep(time.Second)
		if i == 0 {
			err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "LocalSettings.php"), filepath.Join(path, "config", "LocalSettingsBackup.php"))
			if err != nil {
				return canastaInfo, err
			}
		} else {
			err, _ = execute.Run(path, "rm", filepath.Join(path, "config", "LocalSettings.php"))
			if err != nil {
				return canastaInfo, err
			}
		}
		time.Sleep(time.Second)
	}

	if len(WikiNames) == 1 {
		settingsName = "LocalSettings.php"
	}

	err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "LocalSettingsBackup.php"), filepath.Join(path, "config", settingsName))
	if err != nil {
		return canastaInfo, err
	}
	return canastaInfo, err
}

func InstallOne(path, name, domain, wikipath, orchestrator string) error {
	var err error
	logging.Print("Configuring MediaWiki Installation\n")
	logging.Print("Running install.php\n")
	envVariables := canasta.GetEnvVariable(path + "/.env")

	command := "/wait-for-it.sh -t 60 db:3306"
	output, err := orchestrators.ExecWithError(path, orchestrator, "web", command)
	if err != nil {
		return fmt.Errorf(output)
	}

	localExists, _ := fileExists(filepath.Join(path, "config", "LocalSettings.php"))
	commonExists, _ := fileExists(filepath.Join(path, "config", "CommonSettings.php"))

	if !localExists && !commonExists {
		return fmt.Errorf("Neither LocalSettings.php nor CommonSettings.php exist in the path")
	}

	if commonExists {
		err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "CommonSettings.php"), filepath.Join(path, "config", "CommonSettingsBackup.php"))
		if err != nil {
			return err
		}
	}

	if localExists {
		err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "LocalSettings.php"), filepath.Join(path, "config", "LocalSettingsBackup.php"))
		if err != nil {
			return err
		}
	}

	file, err := os.Open(filepath.Join(path, ".admin-password"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // get the first line
	AdminPassword := scanner.Text()

	file, err = os.Open(filepath.Join(path, ".admin"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner = bufio.NewScanner(file)
	scanner.Scan() // get the first line
	AdminName := scanner.Text()

	command = fmt.Sprintf("php maintenance/install.php --dbserver=%s --dbname='%s' --confpath=%s --scriptpath=%s --server='https://%s' --dbuser='%s' --dbpass='%s'  --pass='%s' '%s' '%s'",
		dbServer, name, confPath, scriptPath, domain, "root", envVariables["MYSQL_PASSWORD"], AdminPassword, name, AdminName)
	output, err = orchestrators.ExecWithError(path, orchestrator, "web", command)
	if err != nil {
		return fmt.Errorf(output)
	}

	if localExists {
		err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "LocalSettingsBackup.php"), filepath.Join(path, "config", "CommonSettings.php"))
		if err != nil {
			return err
		}
	}

	if commonExists {
		err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "CommonSettingsBackup.php"), filepath.Join(path, "config", "CommonSettings.php"))
		if err != nil {
			return err
		}
	}

	err, _ = execute.Run(path, "rm", filepath.Join(path, "config", "LocalSettings.php"))
	if err != nil {
		return err
	}

	return err
}

func RemoveDatabase(path, name, orchestrator string) error {
	envVariables := canasta.GetEnvVariable(path + "/.env")
	command := fmt.Sprintf("echo 'DROP DATABASE IF EXISTS %s;' | mysql -h db -u root -p'%s'", name, envVariables["MYSQL_PASSWORD"])
	output, err := orchestrators.ExecWithError(path, orchestrator, "db", command)
	if err != nil {
		return fmt.Errorf("Error while dropping database '%s': %v. Output: %s", name, err, output)
	}

	return nil
}

func passwordCheck(admin, password string) error {
	if len(password) < 10 {
		logging.Fatal(fmt.Errorf("Password must be at least 10 characters long "))
	} else if strings.Contains(password, admin) || strings.Contains(admin, password) {
		logging.Fatal(fmt.Errorf("Password should not be same as admin name"))
	}

	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
