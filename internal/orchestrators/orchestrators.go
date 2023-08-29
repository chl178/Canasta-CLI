package orchestrators

import (
	"fmt"
	"os/exec"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/execute"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/logging"
)

func CheckDependencies() {
	compose := config.GetOrchestrator("docker-compose")
	if compose.Path != "" {
		cmd := exec.Command(compose.Path, "version")
		err := cmd.Run()
		if err != nil {
			logging.Fatal(fmt.Errorf("unable to execute compose (%s) \n", err))
		}
	} else {
		cmd := exec.Command("docker", "compose", "version")
		err := cmd.Run()
		if err != nil {
			logging.Fatal(fmt.Errorf("docker compose should be installed! (%s) \n", err))
		}
	}
}

func GetRepoLink(orchestrator string) string {
	var repo string
	switch orchestrator {
	case "docker-compose":
		repo = "https://github.com/chl178/Canasta-DockerCompose.git"
	default:
		logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
	}
	return repo
}

func CopyOverrideFile(path, orchestrator, sourceFilename, pwd string) error {
	if sourceFilename != "" {
		logging.Print("Copying override file\n")
		switch orchestrator {
		case "docker-compose":
			sourceFilename = pwd + "/" + sourceFilename
			var overrideFilename = path + "/docker-compose.override.yml"
			logging.Print(fmt.Sprintf("Copying %s to %s\n", sourceFilename, overrideFilename))
			err, output := execute.Run("", "cp", sourceFilename, overrideFilename)
			if err != nil {
				logging.Fatal(fmt.Errorf(output))
			}
		default:
			logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
		}
	}
	return nil
}

func Start(path, orchestrator string) error {
	logging.Print("Starting Canasta\n")
	switch orchestrator {
	case "docker-compose":
		compose := config.GetOrchestrator("docker-compose")
		if compose.Path != "" {
			err, output := execute.Run(path, compose.Path, "up", "-d")
			if err != nil {
				return fmt.Errorf(output)
			}
		} else {
			err, output := execute.Run(path, "docker", "compose", "up", "-d")
			if err != nil {
				return fmt.Errorf(output)
			}
		}
	default:
		logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
	}
	return nil
}

func Stop(path, orchestrator string) error {
	logging.Print("Stopping the containers\n")
	switch orchestrator {
	case "docker-compose":
		compose := config.GetOrchestrator("docker-compose")
		if compose.Path != "" {
			err, output := execute.Run(path, compose.Path, "down")
			if err != nil {
				return fmt.Errorf(output)

			}
		} else {
			err, output := execute.Run(path, "docker", "compose", "down")
			if err != nil {
				return fmt.Errorf(output)
			}
		}
	default:
		logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
	}
	return nil
}

func StopAndStart(path, orchestrator string) error {
	if err := Stop(path, orchestrator); err != nil {
		return err
	}
	if err := Start(path, orchestrator); err != nil {
		return err
	}
	return nil
}

func DeleteContainers(path, orchestrator string) (string, error) {
	switch orchestrator {
	case "docker-compose":
		compose := config.GetOrchestrator("docker-compose")
		if compose.Path != "" {

			err, output := execute.Run(path, compose.Path, "down", "-v")
			return output, err
		} else {
			err, output := execute.Run(path, "docker", "compose", "down", "-v")
			return output, err
		}
	default:
		logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
	}
	return "", nil
}

func DeleteConfig(path string) (string, error) {
	//Deleting the installation folder
	err, output := execute.Run("", "rm", "-rf", path)
	return output, err
}

func ExecWithError(path, orchestrator, container, command string) (string, error) {
	var outputByte []byte
	var err error

	switch orchestrator {
	case "docker-compose":
		compose := config.GetOrchestrator("docker-compose")
		if compose.Path != "" {

			cmd := exec.Command(compose.Path, "exec", "-T", container, "/bin/bash", "-c", command)
			if path != "" {
				cmd.Dir = path
			}
			outputByte, err = cmd.CombinedOutput()
		} else {
			cmd := exec.Command("docker", "compose", "exec", "-T", container, "/bin/bash", "-c", command)
			if path != "" {
				cmd.Dir = path
			}
			outputByte, err = cmd.CombinedOutput()
		}
	default:
		logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
	}
	output := string(outputByte)
	logging.Print(output)
	return output, err
}

func Exec(path, orchestrator, container, command string) string {
	output, err := ExecWithError(path, orchestrator, container, command)
	if err != nil {
		logging.Fatal(fmt.Errorf(output))
	}
	return output

}

func CheckRunningStatus(path, id, orchestrator string) error {
	containerName := "web"

	switch orchestrator {
	case "docker-compose":
		compose := config.GetOrchestrator("docker-compose")
		var output string
		var err error
		if compose.Path != "" {
			err, output = execute.Run(path, compose.Path, "ps", "-q", containerName)
		} else {
			err, output = execute.Run(path, "docker", "compose", "ps", "-q", containerName)
		}
		if err != nil || output == "" {
			logging.Fatal(fmt.Errorf("Container %s is not running, please start it first if you want to add a new wiki!", containerName))
			return fmt.Errorf("Container %s is not running", containerName)
		}
	default:
		logging.Fatal(fmt.Errorf("Orchestrator: %s is not available", orchestrator))
		return fmt.Errorf("Orchestrator: %s is not available", orchestrator)
	}
	return nil
}

func ExportDatabase(path, orchestrator, wikiName, outputFilePath string) error {
	// MySQL user, password and container name
	// Replace with your actual MySQL username and password and MySQL container name
	mysqlUser := "root"
	mysqlPassword := "mediawiki"
	mysqlContainerName := "db" // adjust as per your setup

	// Constructing mysqldump command
	dumpCommand := fmt.Sprintf("mysqldump -u %s -p%s %s > /tmp/%s.sql", mysqlUser, mysqlPassword, wikiName, wikiName)
	// Executing mysqldump command inside the MySQL container
	_, err := ExecWithError(path, orchestrator, mysqlContainerName, dumpCommand)
	if err != nil {
		return fmt.Errorf("Failed to execute mysqldump command: %v", err)
	}

	// After executing the mysqldump command, the dump file is inside the container.
	// We need to copy it from the container to the host machine.

	// Constructing docker cp command
	copyCommand := fmt.Sprintf("docker cp %s:/tmp/%s.sql %s", mysqlContainerName, wikiName, outputFilePath)

	// Executing docker cp command on the host machine
	err, output := execute.Run("", "/bin/bash", "-c", copyCommand)
	if err != nil {
		return fmt.Errorf("Failed to copy the dump file from the container: %s", output)
	}

	// Construct the remove command to delete the .sql file from the container
	removeCommand := fmt.Sprintf("rm /tmp/%s.sql", wikiName)

	// Execute the remove command
	_, err = ExecWithError(path, orchestrator, mysqlContainerName, removeCommand)
	if err != nil {
		logging.Fatal(fmt.Errorf("Failed to remove .sql file from container: %w", err))
	}

	return nil
}

func ImportDatabase(databaseName, databasePath string, instance config.Installation) error {
	dbUser := "root"
	dbPassword := "mediawiki"

	// Copy the .sql file into the db container
	copyCmdStr := fmt.Sprintf("docker cp %s db:/tmp/%s.sql", databasePath, databaseName)
	_, err := exec.Command("/bin/bash", "-c", copyCmdStr).Output()
	if err != nil {
		return fmt.Errorf("error copying .sql file to container: %w", err)
	}

	// Ensure the temporary .sql file is removed after the function returns
	defer func() {
		rmCmdStr := fmt.Sprintf("rm /tmp/%s.sql", databaseName)
		_, err := ExecWithError(instance.Path, instance.Orchestrator, "db", rmCmdStr)
		if err != nil {
			logging.Fatal(fmt.Errorf("error removing .sql file from container: %w", err))
		}
	}()

	// Run the mysql command to create the new database
	createCmdStr := fmt.Sprintf("mysql -u%s -p%s -e 'CREATE DATABASE IF NOT EXISTS %s'", dbUser, dbPassword, databaseName)
	_, err = ExecWithError(instance.Path, instance.Orchestrator, "db", createCmdStr)
	if err != nil {
		return fmt.Errorf("error creating database: %w", err)
	}

	// Run the mysql command to import the .sql file into the new database
	importCmdStr := fmt.Sprintf("mysql -u%s -p%s %s < /tmp/%s.sql", dbUser, dbPassword, databaseName, databaseName)
	_, err = ExecWithError(instance.Path, instance.Orchestrator, "db", importCmdStr)
	if err != nil {
		return fmt.Errorf("error importing database: %w", err)
	}

	return nil
}
