package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/urfave/cli"
)

func cliPushNupkg(c *cli.Context) error {

	// Check .nuspec file has been supplied
	filename := c.Args().First()
	if filename == "" {
		return errors.New("Error NU5002: Please specify a nuspec file to use")
	}

	// Read in package contents
	fileContents, err := ioutil.ReadFile(filename)
	checkError(err)

	// print out for CLI
	fmt.Println("Pushing " + filename + " to '" + c.String("Source") + "'...")
	fmt.Println(" ", http.MethodPut, c.String("Source"))
	status, dur, err := PushNupkg(fileContents, c.String("ApiKey"), c.String("Source"))
	checkError(err)

	// print out for CLI
	fmt.Println(" ", status, c.String("Source"), dur, "ms")

	// Log out result
	if status >= 200 && status <= 299 {
		fmt.Println("Your package was pushed.")
	} else {
		fmt.Println("Response status code does not indicate success:", status)
	}

	// Return Ok
	return nil
}

// PushNupkg PUTs a .nupkg binary to a NuGet Repository
func PushNupkg(fileContents []byte, apiKey string, host string) (int, int64, error) {

	// If no Source provided, exit
	if host == "" {
		return 0, 0, errors.New("Error: Please specify a Source/Host")
	}

	// Create MultiPart Writer
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	// Create new File part
	p, err := w.CreateFormFile("package", "package.nupkg")
	checkError(err)
	// Write contents to part
	_, err = p.Write(fileContents)
	checkError(err)
	// Close the writer
	err = w.Close()
	checkError(err)

	// Create new PUT request
	request, err := http.NewRequest(http.MethodPut, host, body)
	checkError(err)
	// Add the ApiKey if supplied
	if apiKey != "" {
		request.Header.Add("X-Nuget-Apikey", apiKey)
	}
	// Add the Content Type header from the reader - includes boundary
	request.Header.Add("Content-Type", w.FormDataContentType())

	// Push to the server
	startTime := time.Now()
	client := &http.Client{}
	resp, err := client.Do(request)
	checkError(err)
	duration := time.Now().Sub(startTime)

	// Return Results
	return resp.StatusCode, duration.Milliseconds(), nil
}
