package execute

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/logging"
)

func Run(path, command string, cmdArgs ...string) (error, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	logging.Print(fmt.Sprint(command, " ", strings.Join(cmdArgs, " ")))
	cmd := exec.Command("bash", "-c", command+" "+strings.Join(cmdArgs, " "))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if path != "" {
		cmd.Dir = path
	}

	if err := cmd.Start(); err != nil {
		logging.Fatal(err)
	}
	err := cmd.Wait()
	output := stdout.String() + stderr.String()
	logging.Print(output)
	return err, output
}
