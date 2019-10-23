package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	nuspec "github.com/soloworks/go-nuspec"
	"github.com/urfave/cli"
)

// Version Number -ldflags="-X 'main.Version=xX.Y.Z'"
var version string = "0.0.0-source"

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
	app.Version = version
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
			},
			Action: packNupkg,
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

func archiveFile(filename string, w *zip.Writer, b []byte) {

	// Create the .nuspec file to the root of the zip
	f, err := w.Create(filename)
	checkError(err)

	// Write .nuspec bytes to file
	_, err = f.Write([]byte(b))
	checkError(err)
}

// Package up a NuSpec file
func packNupkg(c *cli.Context) error {

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

	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive
	w := zip.NewWriter(buf)
	defer w.Close()

	// Create a new Contenttypes Structure
	ct := NewContentTypes()

	// Add .nuspec to Archive
	b, err := n.ToBytes()
	checkError(err)
	archiveFile(filepath.Base(filename), w, b)
	ct.Add(filepath.Ext(filename))

	// Walk the basePath and zip up all found files
	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Base(path) != filepath.Base(filename) {
			x, err := os.Open(path)
			checkError(err)
			y, err := ioutil.ReadAll(x)
			checkError(err)
			p, err := filepath.Rel(basePath, path)
			checkError(err)
			archiveFile(p, w, y)

			ct.Add(filepath.Ext(p))
		}
		return nil
	})

	// Add [Content_Types].xml to Archive
	b, err = ct.ToBytes()
	checkError(err)
	archiveFile(`[Content_Types].xml`, w, b)

	// Create and add .psmdcp file to Archive
	pf := NewPsmdcpFile()
	pf.Creator = n.Meta.Authors
	pf.Description = n.Meta.Description
	pf.Identifier = n.Meta.ID
	pf.Version = n.Meta.Version
	pf.Keywords = n.Meta.Tags
	pf.LastModifiedBy = "go-nuget"
	b, err = pf.ToBytes()
	checkError(err)
	pfn := "/package/services/metadata/core-properties/" + randomString(32) + ".psmdcp"
	archiveFile(pfn, w, b)
	ct.Add(filepath.Ext(pfn))

	// Create and add .rels to Archive
	rf := NewRelFile()
	rf.Add("http://schemas.microsoft.com/packaging/2010/07/manifest", "/"+filepath.Base(filename))
	rf.Add("http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties", pfn)

	b, err = rf.ToBytes()
	checkError(err)
	archiveFile(filepath.Join("_rels", ".rels"), w, b)
	ct.Add(filepath.Ext(".rels"))

	// Close the zipwriter
	w.Close()

	// Ensure directory is present
	if outputPath != "" {
		os.MkdirAll(outputPath, os.ModePerm)
	}
	// Create new file on disk
	outputFile := n.Meta.ID + "." + n.Meta.Version + ".nupkg"
	outputFile = filepath.Join(outputPath, outputFile)
	err = ioutil.WriteFile(outputFile, buf.Bytes(), os.ModePerm)
	checkError(err)

	if err != nil {
		log.Fatal(err)
	}

	println("Successfully created package '" + outputFile + "'")
	return nil
}

func pushNupkg(c *cli.Context) error {

	// Check .nuspec file has been supplied
	filename := c.Args().First()
	if filename == "" {
		return errors.New("Error NU5002: Please specify a nuspec file to use")
	}

	// Create MultiPart Writer
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	// Create new File part
	p, err := w.CreateFormFile("package", "package.nupkg")
	checkError(err)
	// Read in package contents
	fileContents, err := ioutil.ReadFile(filename)
	checkError(err)
	// Write contents to part
	_, err = p.Write(fileContents)
	checkError(err)
	// Close the writer
	err = w.Close()
	checkError(err)

	// Create new PUT request
	request, err := http.NewRequest(http.MethodPut, c.String("Source"), body)
	checkError(err)
	// Add the ApiKey if supplied
	if c.String("ApiKey") != "" {
		request.Header.Add("X-Nuget-Apikey", c.String("ApiKey"))
	}
	// Add the Content Type header from the reader - includes boundary
	request.Header.Add("Content-Type", w.FormDataContentType())

	// Push to the server
	fmt.Println("Pushing " + filename + " to '" + c.String("Source") + "'...")
	fmt.Println(" ", request.Method, request.URL.String())
	startTime := time.Now()
	client := &http.Client{}
	resp, err := client.Do(request)
	checkError(err)
	duration := time.Now().Sub(startTime)

	// Log out result
	fmt.Println(" ", resp.StatusCode, resp.Request.URL.String(), duration.Milliseconds(), "ms")
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		fmt.Println("Your package was pushed.")
	} else {
		fmt.Println("Response status code does not indicate success:", resp.StatusCode)
	}

	return nil
}
