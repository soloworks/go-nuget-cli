package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	nuget "github.com/soloworks/go-nuget-utils"
	nuspec "github.com/soloworks/go-nuspec"
	"github.com/urfave/cli"
)

func cliPackNupkg(c *cli.Context) error {

	// Get filename from command line
	nsfilename := c.Args().First()

	// Check .nuspec file has been supplied
	if nsfilename == "" {
		return errors.New("Error NU5002: Please specify a nuspec file to use")
	}
	// Log out
	fmt.Println("Attempting to build package from '" + nsfilename + "'.")
	// Read in the nuspec file
	ns, err := nuspec.FromFile(nsfilename)
	checkError(err)

	// Set BasePath based on file provided
	basePath := filepath.Dir(nsfilename)
	// Override basePath if argument is provided
	if bp := c.String("BasePath"); bp != "" {
		basePath = bp
		// Check BasePath exists
		if _, err := os.Stat(bp); os.IsNotExist(err) {
			log.Fatalln("Error: BasePath not found")
		}
	}

	// Set OutputDirectory based on file provided
	outputPath := ""
	// Override OutputDirectory if option is set
	if op := c.String("OutputDirectory"); op != "" {
		outputPath = op
	}

	b, err := nuget.PackNupkg(ns, basePath, outputPath)
	checkError(err)

	// Override Version if option is set
	if v := c.String("Version"); v != "" {
		ns.Meta.Version = v
	}

	// Ensure directory is present
	if outputPath != "" {
		os.MkdirAll(outputPath, os.ModePerm)
	}
	// Create new file on disk
	outputFile := ns.Meta.ID + "." + ns.Meta.Version + ".nupkg"
	outputFile = filepath.Join(outputPath, outputFile)
	err = ioutil.WriteFile(outputFile, b, os.ModePerm)
	checkError(err)

	if err != nil {
		log.Fatal(err)
	}

	println("Successfully created package '" + outputFile + "'")

	return nil
}
