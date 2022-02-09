package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

var (
	Version    = "snapshot"
	FullCommit = "FullCommit"
	GitURL     = "GitURL"
)

func er(err error) {
	if err != nil {
		fmt.Println("There was an error trying to execute your request:")
		fmt.Println(err)
		os.Exit(1)
	}
}

// mustAsk simplifies command code since removes duplicate code
func mustAsk(c *cli.Context, flagName, question, defaultAnswer string, isPassword bool, validationFunc promptui.ValidateFunc) string {

	var (
		value string
		err   error
	)

	value = c.String(flagName)

	if value == "" {
		if isPassword {
			value, err = promptPassword(question, validationFunc)
		} else {
			value, err = promptText(question, defaultAnswer, validationFunc)
		}
	}

	er(err)
	return value

}

// promptText formats a prompt and returns its result
func promptText(promptLabel, promptDefault string, validateFunc promptui.ValidateFunc) (string, error) {

	var (
		prompt promptui.Prompt
	)

	prompt.Label = promptLabel
	if promptDefault != "" {
		prompt.Default = promptDefault
	}

	if validateFunc != nil {
		prompt.Validate = validateFunc
	}

	return prompt.Run()

}

// promptPassword formats a prompt for password request, returns its result and validation error if any
func promptPassword(promptLabel string, validateFunc promptui.ValidateFunc) (string, error) {

	var (
		prompt promptui.Prompt
	)

	prompt.Label = promptLabel
	prompt.Mask = '*'
	if validateFunc != nil {
		prompt.Validate = validateFunc
	}

	return prompt.Run()

}

// promptTrueFalseBool returns a prompt for true, false questions
// -  promptLabel: Text to use as label
// -   trueString: Text for true
// -  falseString: Text for false
// - defaultValue: default value
func promptTrueFalseBool(promptLabel, trueString, falseString string, defaultValue bool) (bool, error) {

	var (
		items  []string
		result string
		err    error
	)

	if defaultValue {
		items = []string{trueString, falseString}
	} else {
		items = []string{falseString, trueString}
	}

	prompt := promptui.Select{
		Label: promptLabel,
		Items: items,
	}

	_, result, err = prompt.Run()
	if err != nil {
		return false, err
	}

	if result == trueString {
		return true, err
	}
	return false, err

}

var errRequired = errors.New("input required")

// validationRequired validates the string is not empty
func validationRequired(input string) error {

	if input == "" {
		return errRequired
	}

	return nil

}
