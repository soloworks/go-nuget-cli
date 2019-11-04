package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	nuget "github.com/soloworks/go-nuget-utils"
	"github.com/urfave/cli"
)

func cliSampleNuSpec(c *cli.Context) error {

	// Get current user details
	user, err := user.Current()
	checkError(err)

	// Get the NuSpec
	ns := nuget.SampleNuSpec(c.Args().First(), user.Name)

	// Set filename string
	fn := ns.Meta.ID + ".nuspec"

	// Check if file exists and -Force isn't active
	if _, err := os.Stat(fn); !os.IsNotExist(err) {
		if !c.Bool("Force") {
			return errors.New("'" + fn + "' already exists, use -Force to overwrite it.")
		}
	}

	// Convert to []byte
	b, err := ns.ToBytes()
	checkError(err)

	// Write to filesystem
	err = ioutil.WriteFile(fn, b, 0644)
	checkError(err)

	// Echo out message
	fmt.Println("Created: '" + fn + "' successfully.")

	return nil
}
