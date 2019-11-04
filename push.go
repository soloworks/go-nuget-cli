package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	nuget "github.com/soloworks/go-nuget-utils"
	"github.com/urfave/cli"
)

func cliPushNupkg(c *cli.Context) error {

	// Check .nuspec file has been supplied
	filename := c.Args().First()
	if filename == "" {
		return errors.New("Error NU5002: Please specify a nuspec file to use")
	}

	// Read in package contents
	fileContents, err := ioutil.ReadFile(filename)
	checkError(err)

	// print out for CLI
	fmt.Println("Pushing " + filename + " to '" + c.String("Source") + "'...")
	fmt.Println(" ", http.MethodPut, c.String("Source"))
	status, dur, err := nuget.PushNupkg(fileContents, c.String("ApiKey"), c.String("Source"))
	checkError(err)

	// print out for CLI
	fmt.Println(" ", status, c.String("Source"), dur, "ms")

	// Log out result
	if status >= 200 && status <= 299 {
		fmt.Println("Your package was pushed.")
	} else {
		fmt.Println("Response status code does not indicate success:", status)
	}

	// Return Ok
	return nil
}
