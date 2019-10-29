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
