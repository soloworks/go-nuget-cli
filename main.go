package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"time"

	nuspec "github.com/soloworks/go-nuspec"
	"github.com/urfave/cli"
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	app := cli.NewApp()
	app.Name = "go-nuget"
	app.Usage = "An open source nuget clone in Go"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Sam Shelton",
			Email: "sam.shelton@soloworks.co.uk",
		},
	}
	app.Copyright = "(c) 2019 Solo Works London"

	// Subcommands
	app.Commands = []cli.Command{
		{
			Name:    "pack",
			Aliases: []string{"a"},
			Usage:   "generate .nupkg file from .nuspec",
			Action: func(c *cli.Context) error {
				fmt.Println("added task: ", c.Args().First())
				return nil
			},
		},
		{
			Name:    "push",
			Aliases: []string{"c"},
			Usage:   "upload a .nupkg file to a nuget server",
			Action: func(c *cli.Context) error {
				fmt.Println("completed task: ", c.Args().First())
				return nil
			},
		},
		{
			Name:    "spec",
			Aliases: []string{"c"},
			Usage:   "Generates a nuspec for a new package.",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "Force",
					Usage: "Overwrite nuspec file if it exists",
				},
			},
			Action: sampleNuspec,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func sampleNuspec(c *cli.Context) error {

	// Get current user details
	user, err := user.Current()
	checkError(err)

	// Create a new structure
	n := nuspec.New()

	// Populate Defaults
	n.Meta.ID = "Package"
	n.Meta.Version = "1.0.0"
	n.Meta.Authors = user.Name
	n.Meta.Owners = user.Name
	n.Meta.LicenseURL = "http://LICENSE_URL_HERE_OR_DELETE_THIS_LINE"
	n.Meta.ProjectURL = "http://LICENSE_URL_HERE_OR_DELETE_THIS_LINE"
	n.Meta.IconURL = "http://LICENSE_URL_HERE_OR_DELETE_THIS_LINE"
	n.Meta.ReqLicenseAccept = false
	n.Meta.Description = "Package Description"
	n.Meta.ReleaseNotes = "Summary of changes made in this release of the package."
	n.Meta.Copyright = "Copyright " + time.Now().Format("2006")
	n.Meta.Tags = "Tag1 Tag2"
	d := nuspec.Dependency{ID: "SampleDependency", Version: "1.0"}
	n.Meta.Dependencies.Dependency = append(n.Meta.Dependencies.Dependency, d)

	// Override package ID if present
	if c.Args().First() != "" {
		n.Meta.ID = c.Args().First()
	}

	// Output it to the disk
	fn := n.Meta.ID + ".nuspec"

	if _, err := os.Stat(fn); !os.IsNotExist(err) {
		if !c.Bool("Force") {
			println("'" + fn + "' already exists, use -Force to overwrite it.")
			os.Exit(1)
		}
	}

	b, err := n.ToBytes()
	checkError(err)
	err = ioutil.WriteFile(fn, b, 0644)
	checkError(err)
	// Echo out message
	fmt.Println("Created: '" + fn + "' successfully.")
	return nil
}
