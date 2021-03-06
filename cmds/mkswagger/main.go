package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/go-swagger/go-swagger/cmd/swagger/commands/generate"
	"github.com/go-swagger/go-swagger/scan"
	flags "github.com/jessevdk/go-flags"
	bindata "github.com/kevinburke/go-bindata"
)

var gofiles *regexp.Regexp

var opts struct {
	OutputPackage flags.Filename `long:"package" short:"p" description:"destination package" env:"GOPACKAGE"`
	Input         flags.Filename `long:"input" short:"i" description:"go package to use as input" required:"t"`
}

func init() {
	gofiles = regexp.MustCompile(`.+\.go$`)
}

func generateSpec(s generate.SpecFile, outputFilename string) error {
	var opts scan.Opts
	opts.BasePath = s.BasePath
	opts.ScanModels = s.ScanModels
	opts.BuildTags = s.BuildTags
	swspec, err := scan.Application(opts)
	if err != nil {
		return err
	}

	var b []byte
	b, err = json.MarshalIndent(swspec, "", "  ")

	if err != nil {
		return err
	}
	if outputFilename == "" {
		fmt.Println(string(b))
		return nil
	}
	return ioutil.WriteFile(outputFilename, b, 0644)
}

func main() {

	pwd, _ := os.Getwd()
	_, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(1)
	}

	bdConfig := bindata.Config{
		Package:    string(opts.OutputPackage),
		Output:     path.Join(pwd, "bindata.go"),
		Prefix:     pwd,
		NoMemCopy:  true,
		NoMetadata: true,
		Ignore:     []*regexp.Regexp{gofiles},
		Input: []bindata.InputConfig{
			bindata.InputConfig{
				Path:      pwd,
				Recursive: true,
			},
		},
	}
	// generate bindata to bootstrap, if necessary
	err = bindata.Translate(&bdConfig)
	if err != nil {
		fmt.Println("Error bootstrapping bindata.go: ", err)
	}

	// generate the swagger specification
	outputFilename := path.Join(pwd, "swagger.json")
	fmt.Println("Generating swagger.json...")
	spec := generate.SpecFile{
		ScanModels: true,
		BasePath:   string(opts.Input),
		Compact:    false,
	}
	err = generateSpec(spec, outputFilename)
	if err != nil {
		// Note that go-swagger has the annoying habit of Fatal-ing,
		// so we may have already exit-ed before reaching this point.
		fmt.Println("Error generating swagger.json: ", err)
	}

	// regenerate bindata.go
	fmt.Println("Generating bindata.go...")
	err = bindata.Translate(&bdConfig)
	if err != nil {
		fmt.Println("Error generating bindata.go: ", err)
	}
}
