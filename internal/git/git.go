package git

import (
	"fmt"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/execute"
)

func Clone(repo, path string) error {
	err, output := execute.Run("", "git", "clone", repo, path)
	if err != nil {
		return fmt.Errorf(output)
	}
	return nil
}
func Cloneb(repo, path string, branch string) error {
	err, output := execute.Run("", "git", "clone", "-b", branch, repo, path)
	if err != nil {
		return fmt.Errorf(output)
	}
	return nil
}
