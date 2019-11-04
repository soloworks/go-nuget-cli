package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli"
)

// Version Number -ldflags="-X 'main.Version=xX.Y.Z'"
var version string = "0.0.0.0-master"
var compiled string = "1262304000" // 01/01/2010 @ 12:00am

const letterBytes = "abcdef0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func checkError(e error) {
	if e != nil {
		println(e.Error())
		os.Exit(1)
	}
}

func main() {
	app := cli.NewApp()
	// Build Values - Version String
	app.Version = version
	// Build Values - Compiled Timestamp from Unix Time String
	i, err := strconv.ParseInt(compiled, 10, 64)
	if err != nil {
		panic(err)
	}
	app.Compiled = time.Unix(i, 0)

	app.Name = "go-nuget"
	app.Usage = "An open source nuget clone in Go"
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
				cli.StringFlag{
					Name:  "Version",
					Usage: "Overrides the version number from the nuspec file.",
				},
			},
			Action: cliPackNupkg,
		},
		{
			Name:  "push",
			Usage: "Pushes a package to the server and publishes it.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "ApiKey",
					Usage: "The API key for the target repository.",
				},
				cli.StringFlag{
					Name:  "Source",
					Usage: "Specifies the server URL.",
				},
			},
			Action: pushNupkg,
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
			Action: cliSampleNuSpec,
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
