package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	nuspec "github.com/soloworks/go-nuspec"
	"github.com/urfave/cli"
)

func archiveFile(fn string, w *zip.Writer, b []byte) {

	// Check and convert filepath to `/` if required
	fn = filepath.ToSlash(fn)

	// Create the file in the zip
	f, err := w.Create(fn)
	checkError(err)

	// Write .nuspec bytes to file
	_, err = f.Write([]byte(b))
	checkError(err)
}

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

	b, err := PackNupkg(ns, basePath, outputPath)
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

// PackNupkg produces a .nupkg file in byte format
func PackNupkg(ns *nuspec.NuSpec, basePath string, outputPath string) ([]byte, error) {

	// Assume filename from ID
	nsfilename := ns.Meta.ID + ".nuspec"

	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive
	w := zip.NewWriter(buf)
	defer w.Close()

	// Create a new Contenttypes Structure
	ct := NewContentTypes()

	// Add .nuspec to Archive
	b, err := ns.ToBytes()
	checkError(err)
	archiveFile(filepath.Base(nsfilename), w, b)
	ct.Add(filepath.Ext(nsfilename))

	// Process files
	// If there are no file globs specified then
	if len(ns.Files.File) == 0 {
		// walk the basePath and zip up all found files. Everything.]
		err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && filepath.Base(path) != filepath.Base(nsfilename) {
				// Open the file
				x, err := os.Open(path)
				checkError(err)
				// Gather all contents
				y, err := ioutil.ReadAll(x)
				checkError(err)
				// Set relative path for file in archive
				p, err := filepath.Rel(basePath, path)
				checkError(err)
				// Store the file
				archiveFile(p, w, y)
				// Add extension to the Rels file
				ct.Add(filepath.Ext(p))
			}
			return nil
		})
		checkError(err)
	} else {
		// For each of the specified globs, get files an put in target
		for _, f := range ns.Files.File {
			// Apply glob, cater for
			matches, err := filepath.Glob(filepath.ToSlash(filepath.Join(basePath, f.Source)))
			checkError(err)
			for _, m := range matches {
				info, err := os.Stat(m)
				if !info.IsDir() && filepath.Base(m) != filepath.Base(nsfilename) {
					// Open the file
					x, err := os.Open(m)
					checkError(err)
					// Gather all contents
					y, err := ioutil.ReadAll(x)
					checkError(err)
					// Set relative path for file in archive
					p, err := filepath.Rel(basePath, m)
					checkError(err)
					// Overide path if Target is set
					if f.Target != "" {
						p = filepath.Join(f.Target, filepath.Base(m))
					}
					// Store the file
					archiveFile(p, w, y)
					// Add extension to the Rels file
					ct.Add(filepath.Ext(p))
				}
				checkError(err)
			}
		}
	}

	// Create and add .psmdcp file to Archive
	pf := NewPsmdcpFile()
	pf.Creator = ns.Meta.Authors
	pf.Description = ns.Meta.Description
	pf.Identifier = ns.Meta.ID
	pf.Version = ns.Meta.Version
	pf.Keywords = ns.Meta.Tags
	pf.LastModifiedBy = "go-nuget"
	b, err = pf.ToBytes()
	checkError(err)
	pfn := "package/services/metadata/core-properties/" + randomString(32) + ".psmdcp"
	archiveFile(pfn, w, b)
	ct.Add(filepath.Ext(pfn))

	// Create and add .rels to Archive
	rf := NewRelFile()
	rf.Add("http://schemas.microsoft.com/packaging/2010/07/manifest", "/"+filepath.Base(nsfilename))
	rf.Add("http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties", pfn)

	b, err = rf.ToBytes()
	checkError(err)
	archiveFile(filepath.Join("_rels", ".rels"), w, b)
	ct.Add(filepath.Ext(".rels"))

	// Add [Content_Types].xml to Archive
	b, err = ct.ToBytes()
	checkError(err)
	archiveFile(`[Content_Types].xml`, w, b)

	// Close the zipwriter
	w.Close()

	// Return
	return buf.Bytes(), nil
}
