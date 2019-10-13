package nuspec

import (
	"bytes"
	"encoding/xml"
)

// File Represents a .nuspec XML file found in the root of the .nupck files
type File struct {
	Package struct {
		Xmlns string `xml:"xmlns"`
	} `xml:"package"`
	Metadata struct {
		ID                       string `xml:"id"`
		Version                  string `xml:"version"`
		Title                    string `xml:"title"`
		Authors                  string `xml:"authors"`
		Owners                   string `xml:"owners"`
		ProjectURL               string `xml:"projectUrl"`
		LicenseURL               string `xml:"licenseUrl"`
		IconURL                  string `xml:"iconUrl"`
		RequireLicenseAcceptance string `xml:"requireLicenseAcceptance"`
		Description              string `xml:"description"`
		ReleaseNotes             string `xml:"releaseNotes"`
		Copyright                string `xml:"copyright"`
		Summary                  string `xml:"summary"`
		Language                 string `xml:"language"`
		Tags                     string `xml:"tags"`
	} `xml:"metadata"`
}

// New returns a populated skeleton for a Nuget Packages request (/Packages)
func New() *File {
	nsf := File{}
	nsf.Package.Xmlns = `http://schemas.microsoft.com/packaging/2010/07/nuspec.xsd`
	return &nsf
}

// FromFile pulls in a nuspec file drom the drive
func FromFile(fn string) (*File, error) {
	nsf := File{}
	return &nsf, nil
}

// FromBytes pulls in a nuspec file drom the drive
func FromBytes(b []byte) (*File, error) {
	nsf := File{}
	err := xml.Unmarshal(b, &nsf)
	if err != nil {
		return nil, err
	}
	return &nsf, nil
}

// ToBytes produces the nuspec in XML format
func (nsf *File) ToBytes() ([]byte, error) {
	var b bytes.Buffer
	// Unmarshal into XML
	output, err := xml.MarshalIndent(nsf, "  ", "    ")
	if err != nil {
		return nil, err
	}
	// Write the XML Header
	b.WriteString(xml.Header)
	b.Write(output)
	return b.Bytes(), nil
}
