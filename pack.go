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

	// Override Version if option is set
	if v := c.String("Version"); v != "" {
		n.Meta.Version = v
	}

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

	// Process files
	// If there are no file globs specified then
	if len(n.Files.File) == 0 {
		// walk the basePath and zip up all found files. Everything.]
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
		checkError(err)
	} else {
		// For each of the specified globs, get files an put in target
		for _, f := range n.Files.File {
			matches, err := filepath.Glob(f.Source)
			checkError(err)
			for _, m := range matches {
				info, err := os.Stat(m)
				if !info.IsDir() && filepath.Base(m) != filepath.Base(filename) {
					x, err := os.Open(m)
					checkError(err)
					y, err := ioutil.ReadAll(x)
					checkError(err)
					p, err := filepath.Rel(basePath, m)
					checkError(err)
					archiveFile(p, w, y)
					ct.Add(filepath.Ext(p))
				}
				checkError(err)
			}
		}
	}

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
