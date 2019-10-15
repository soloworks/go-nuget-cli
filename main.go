package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	nuspec "github.com/soloworks/go-nuspec"
	"github.com/urfave/cli"
)

func checkError(e error) {
	if e != nil {
		println(e.Error())
		os.Exit(1)
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
			Name:  "pack",
			Usage: "Creates a NuGet package based on the specified nuspec file.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "BasePath",
					Usage: "Sets the base path of the files defined in the .nuspec file.",
				},
				cli.StringFlag{
					Name:  "OutputDirectory",
					Usage: "Specifies the folder in which the created package is stored. If no folder is specified, the current folder is used.",
				},
			},
			Action: packNuspec,
		},
		{
			Name:  "push",
			Usage: "Pushes a package to the server and publishes it.",
			Action: func(c *cli.Context) error {
				fmt.Println("completed task: ", c.Args().First())
				return nil
			},
		},
		{
			Name:  "spec",
			Usage: "Generates a nuspec for a new package.",
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

	// Set filename string
	fn := n.Meta.ID + ".nuspec"

	// Check if file exists and -Force isn't active
	if _, err := os.Stat(fn); !os.IsNotExist(err) {
		if !c.Bool("Force") {
			return errors.New("'" + fn + "' already exists, use -Force to overwrite it.")
		}
	}

	// Convert to []byte
	b, err := n.ToBytes()
	checkError(err)

	// Writ to filesystem
	err = ioutil.WriteFile(fn, b, 0644)
	checkError(err)

	// Echo out message
	fmt.Println("Created: '" + fn + "' successfully.")
	return nil
}

// Package up a NuSpec file
func packNuspec(c *cli.Context) error {

	filename := c.Args().First()

	// Check .nuspec file has been supplied
	if filename == "" {
		return errors.New("Error NU5002: Please specify a nuspec file to use")
	}
	// Log out
	fmt.Println("Attempting to build package from '" + filename + "'.")
	// Read in the nuspec file
	n, err := nuspec.FromFile(filename)
	checkError(err)

	// Set BasePath based on file provided
	basePath := filepath.Dir(filename)

	// Override basePath if option is set
	if bp := c.String("BasePath"); bp != "" {
		basePath = bp
	}

	// Set OutputDirectory based on file provided
	outputPath := ""

	// Override OutputDirectory if option is set
	if op := c.String("OutputDirectory"); op != "" {
		outputPath = op
	}

	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)
	defer w.Close()

	// Create the .nuspec file to the root of the zip
	f, err := w.Create(filepath.Base(filename))
	checkError(err)

	// Export .nuspec as bytes
	b, err := n.ToBytes()
	checkError(err)

	// Write .nuspec bytes to file
	_, err = f.Write([]byte(b))
	checkError(err)

	// Walk the basePath and zip up all found files
	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Base(path) != filepath.Base(filename) {
			x, err := os.Open(path)
			checkError(err)
			y, err := ioutil.ReadAll(x)
			checkError(err)
			z, err := w.Create(filepath.Clean(strings.Replace(path, basePath, ".", 1)))
			checkError(err)
			z.Write(y)
			checkError(err)
		}
		return nil
	})

	// Close the zipwriter
	w.Close()

	// Ensure directory is present
	if outputPath != "" {
		os.MkdirAll(outputPath, os.ModePerm)
	}
	// Create new file on disk
	outputFile := n.Meta.ID + "." + n.Meta.Version + ".zip"
	outputFile = filepath.Join(outputPath, outputFile)
	err = ioutil.WriteFile(outputFile, buf.Bytes(), os.ModePerm)
	checkError(err)

	if err != nil {
		log.Fatal(err)
	}

	println("Successfully created package '" + outputFile + "'")
	return nil
}
