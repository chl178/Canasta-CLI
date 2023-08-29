package prompt

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"golang.org/x/term"
)

func PromptUser(name, yamlPath string, canastaInfo canasta.CanastaVariables) (string, canasta.CanastaVariables, error) {
	var err error
	if yamlPath == "" {
		if name, err = promptForInput(name, "WikiID"); err != nil {
			return name, canastaInfo, err
		}
	}
	if canastaInfo.Id, err = promptForInput(canastaInfo.Id, "Canasta ID"); err != nil {
		return name, canastaInfo, err
	}
	if canastaInfo.AdminName, canastaInfo.AdminPassword, err = promptForUserPassword(canastaInfo.AdminName, canastaInfo.AdminPassword); err != nil {
		return name, canastaInfo, err
	}
	return name, canastaInfo, nil
}

func PromptWiki(name, urlString, id, siteName string) (string, string, string, string, string, error) {
	var err error
	if id, err = promptForInput(id, "CanastaID"); err != nil {
		return "", "", "", "", "", err
	}
	if name, err = promptForInput(name, "wikiID"); err != nil {
		return "", "", "", "", "", err
	}
	if siteName, err = promptForInputWithNull(siteName, "site name"); err != nil {
		return "", "", "", "", "", err
	}
	if urlString, err = promptForInput(urlString, "URL"); err != nil {
		return "", "", "", "", "", err
	}

	// add 'http://' to urlString if no schema is present
	if !strings.HasPrefix(urlString, "http://") && !strings.HasPrefix(urlString, "https://") {
		urlString = "https://" + urlString
	}

	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	domain := parsedUrl.Hostname()
	path := strings.Trim(parsedUrl.Path, "/") // remove leading and trailing slashes

	return name, domain, path, id, siteName, nil
}

func promptForInput(value, prompt string) (string, error) {
	if value != "" {
		return value, nil
	}
	return getUserInput(fmt.Sprintf("Enter %s: ", prompt), false)
}

func promptForInputWithNull(value, prompt string) (string, error) {
	if value != "" {
		return value, nil
	}
	return getUserInput(fmt.Sprintf("Enter %s: ", prompt), true)
}

func getUserInput(message string, allowNull bool) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(message)
	scanner.Scan()
	input := scanner.Text()
	if input == "" && !allowNull {
		return "", fmt.Errorf("please enter a value")
	}
	return input, nil
}

func promptForUserPassword(username, password string) (string, string, error) {
	var err error
	if username, err = promptForInput(username, "admin name"); err != nil {
		return "", "", err
	}
	if password == "" {
		if username, password, err = getAndConfirmPassword(username); err != nil {
			return "", "", err
		}
	}
	return username, password, nil
}

func getAndConfirmPassword(username string) (string, string, error) {
	fmt.Print("Enter the admin password (Press Enter to autogenerate the password): \n")
	password, err := getPasswordInput()
	if err != nil {
		return "", "", err
	}
	if password == "" {
		return username, "", nil
	}
	fmt.Print("Re-enter the admin password: \n")
	confirmedPassword, err := getPasswordInput()
	if err != nil || password != confirmedPassword {
		return "", "", fmt.Errorf("Passwords do not match, please try again.")
	}
	return username, password, nil
}

func getPasswordInput() (string, error) {
	bytePassword, err := term.ReadPassword(0)
	return string(bytePassword), err
}
