package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"time"

	nuspec "github.com/soloworks/go-nuspec"
	"github.com/urfave/cli"
)

func cliSampleNuSpec(c *cli.Context) error {

	// Get current user details
	user, err := user.Current()
	checkError(err)

	// Get the NuSpec
	ns := SampleNuSpec(c.Args().First(), user.Name)

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

// SampleNuSpec returns a populated sample NuSpec struct
func SampleNuSpec(id string, user string) *nuspec.NuSpec {

	// Create a new structure
	n := nuspec.New()

	// Populate Defaults
	n.Meta.ID = "Package"
	n.Meta.Version = "1.0.0"
	n.Meta.Authors = user
	n.Meta.Owners = user
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
	if id != "" {
		n.Meta.ID = id
	}

	return n
}
