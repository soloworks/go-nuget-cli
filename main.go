package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	nuspec "github.com/soloworks/go-nuspec"
	"github.com/urfave/cli"
)

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

func archiveFile(filename string, w *zip.Writer, b []byte) {

	// Create the .nuspec file to the root of the zip
	f, err := w.Create(filename)
	checkError(err)

	// Write .nuspec bytes to file
	_, err = f.Write([]byte(b))
	checkError(err)
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

// ContentTypeEntry is used by the ContentTypes struct
type ContentTypeEntry struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// ContentTypes is represents a [Content_Types].xml file from a .nupkg file
type ContentTypes struct {
	XMLName xml.Name            `xml:"Types"`
	Xmlns   string              `xml:"xmlns,attr"`
	Entry   []*ContentTypeEntry `xml:"Default"`
}

// NewContentTypes is a constructor for the ContentTypes struct
func NewContentTypes() *ContentTypes {
	ct := &ContentTypes{}
	ct.Xmlns = "http://schemas.openxmlformats.org/package/2006/content-types"
	return ct
}

// Add pushes a new extension into a ContentType struct
func (ct *ContentTypes) Add(ext string) {
	if strings.HasPrefix(ext, ".") {
		ext = strings.TrimLeft(ext, ".")
	}
	// Create a new entry
	cte := &ContentTypeEntry{Extension: ext}
	// If it already exists we can exit
	for _, e := range ct.Entry {
		if e.Extension == cte.Extension {
			return
		}
	}
	// Set the content type
	switch cte.Extension {
	case "rels":
		cte.ContentType = "application/vnd.openxmlformats-package.relationships+xml"
	case "psmdcp":
		cte.ContentType = "application/vnd.openxmlformats-package.core-properties+xml"
	default:
		cte.ContentType = "application/octet"
	}
	// Add it to the array
	ct.Entry = append(ct.Entry, cte)
}

// ToBytes produces the nuspec in XML format
func (ct *ContentTypes) ToBytes() ([]byte, error) {
	var b bytes.Buffer
	// Unmarshal into XML
	output, err := xml.MarshalIndent(ct, "", "  ")
	if err != nil {
		return nil, err
	}
	// Self-Close any empty XML elements (to match original Nuget output)
	// This assumes Indented Marshalling above, non Indented will break XML
	for bytes.Contains(output, []byte(`></`)) {
		i := bytes.Index(output, []byte(`></`))
		j := bytes.Index(output[i+1:], []byte(`>`))
		output = append(output[:i], append([]byte(` /`), output[i+j+1:]...)...)
	}
	// Write the XML Header
	b.WriteString(xml.Header)
	b.Write(output)
	return b.Bytes(), nil
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

// Rel represents a relationship used in RelFile
type Rel struct {
	Type   string `xml:"Type,attr"`
	Target string `xml:"Target,attr"`
	ID     string `xml:"Id,attr"`
}

// RelFile represents a Relationship File stored in .rels
type RelFile struct {
	XMLName xml.Name `xml:"Relationships"`
	XMLns   string   `xml:"xmlns,attr"`
	Rels    []*Rel   `xml:"Relationship"`
}

// Add appends a new relationship to the list
func (rf *RelFile) Add(t string, targ string) {
	r := &Rel{
		Type:   t,
		Target: targ,
	}
	rf.Rels = append(rf.Rels, r)
	// Add UID (Unique in this file...)
	rf.Rels[len(rf.Rels)-1].ID = fmt.Sprintf("R%015d", len(rf.Rels))
}

// NewRelFile returns a populated skeleton for a Nuget Packages Entry
func NewRelFile() *RelFile {
	// Create new entry
	rf := &RelFile{
		XMLns: "http://schemas.openxmlformats.org/package/2006/relationships",
	}
	return rf
}

// ToBytes exports structure as byte array
func (rf *RelFile) ToBytes() ([]byte, error) {
	var b bytes.Buffer
	// Unmarshal into XML
	output, err := xml.MarshalIndent(rf, "", "  ")
	if err != nil {
		return nil, err
	}
	// Self-Close any empty XML elements (NuGet client is broken and requires this on some)
	// This assumes Indented Marshalling above, non Indented will break XML
	// Break XML Encoding to match Nuget server output
	for bytes.Contains(output, []byte(`></`)) {
		i := bytes.Index(output, []byte(`></`))
		j := bytes.Index(output[i+1:], []byte(`>`))
		output = append(output[:i], append([]byte(` /`), output[i+j+1:]...)...)
	}

	// Write the XML Header
	b.WriteString(xml.Header)
	b.Write(output)
	return b.Bytes(), nil

}

// PsmdcpFile is a variation XML generated by nuget
type PsmdcpFile struct {
	XMLName        xml.Name `xml:"coreProperties"`
	XMLNSdc        string   `xml:"xmlns:dc,attr"`
	XMLNSdcterms   string   `xml:"xmlns:dcterms,attr"`
	XMLNSxsi       string   `xml:"xmlns:xsi,attr"`
	XMLNS          string   `xml:"xmlns,attr"`
	Creator        string   `xml:"dc:creator"`
	Description    string   `xml:"dc:description"`
	Identifier     string   `xml:"dc:identifier"`
	Version        string   `xml:"version"`
	Keywords       string   `xml:"keywords"`
	LastModifiedBy string   `xml:"lastModifiedBy"`
}

// NewPsmdcpFile returns a populated skeleton for a Nuget Packages Entry
func NewPsmdcpFile() *PsmdcpFile {
	// Create new entry
	pf := &PsmdcpFile{
		XMLNSdc:      "http://purl.org/dc/elements/1.1/",
		XMLNSdcterms: "http://purl.org/dc/terms/",
		XMLNSxsi:     "http://www.w3.org/2001/XMLSchema-instance",
		XMLNS:        "http://schemas.openxmlformats.org/package/2006/metadata/core-properties",
	}
	return pf
}

// ToBytes exports structure as byte array
func (pf *PsmdcpFile) ToBytes() ([]byte, error) {
	var b bytes.Buffer
	// Unmarshal into XML
	output, err := xml.MarshalIndent(pf, "", "  ")
	if err != nil {
		return nil, err
	}
	// Self-Close any empty XML elements (NuGet client is broken and requires this on some)
	// This assumes Indented Marshalling above, non Indented will break XML
	// Break XML Encoding to match Nuget server output
	for bytes.Contains(output, []byte(`></`)) {
		i := bytes.Index(output, []byte(`></`))
		j := bytes.Index(output[i+1:], []byte(`>`))
		output = append(output[:i], append([]byte(` /`), output[i+j+1:]...)...)
	}

	// Write the XML Header
	b.WriteString(xml.Header)
	b.Write(output)
	return b.Bytes(), nil
}
