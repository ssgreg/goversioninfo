// Contribution by Tamás Gulácsi

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/ssgreg/goversioninfo"
)

func main() {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)

	flagExample := flagSet.Bool("example", false, "just dump out an example versioninfo.json to stdout")
	flagOut := flagSet.String("o", "resource.syso", "output file name")
	flagPlatformSpecific := flagSet.Bool("platform-specific", false, "output i386 and amd64 named resource.syso, ignores -o")
	flagIcon := flagSet.String("icon", "", "icon file name")
	flagManifest := flagSet.String("manifest", "", "manifest file name")

	flagComment := flagSet.String("comment", "", "StringFileInfo.Comments")
	flagCompany := flagSet.String("company", "", "StringFileInfo.CompanyName")
	flagDescription := flagSet.String("description", "", "StringFileInfo.FileDescription")
	flagFileVersion := flagSet.String("file-version", "", "StringFileInfo.FileVersion")
	flagInternalName := flagSet.String("internal-name", "", "StringFileInfo.InternalName")
	flagCopyright := flagSet.String("copyright", "", "StringFileInfo.LegalCopyright")
	flagTrademark := flagSet.String("trademark", "", "StringFileInfo.LegalTrademarks")
	flagOriginalName := flagSet.String("original-name", "", "StringFileInfo.OriginalFilename")
	flagPrivateBuild := flagSet.String("private-build", "", "StringFileInfo.PrivateBuild")
	flagProductName := flagSet.String("product-name", "", "StringFileInfo.ProductName")
	flagProductVersion := flagSet.String("product-version", "", "StringFileInfo.ProductVersion")
	flagSpecialBuild := flagSet.String("special-build", "", "StringFileInfo.SpecialBuild")

	flagTranslation := flagSet.Int("translation", 0, "translation ID")
	flagCharset := flagSet.Int("charset", 0, "charset ID")

	flag64 := flagSet.Bool("64", false, "generate 64-bit binaries")

	flagVerMajor := flagSet.Int("ver-major", -1, "FileVersion.Major")
	flagVerMinor := flagSet.Int("ver-minor", -1, "FileVersion.Minor")
	flagVerPatch := flagSet.Int("ver-patch", -1, "FileVersion.Patch")
	flagVerBuild := flagSet.Int("ver-build", -1, "FileVersion.Build")

	flagProductVerMajor := flagSet.Int("product-ver-major", -1, "ProductVersion.Major")
	flagProductVerMinor := flagSet.Int("product-ver-minor", -1, "ProductVersion.Minor")
	flagProductVerPatch := flagSet.Int("product-ver-patch", -1, "ProductVersion.Patch")
	flagProductVerBuild := flagSet.Int("product-ver-build", -1, "ProductVersion.Build")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <versioninfo.json>\n\nPossible flags:\n", os.Args[0])
		flagSet.PrintDefaults()
	}
	var err error
	flagSet.Parse(os.Args[2:])
	if *flagExample {
		io.WriteString(os.Stdout, example)
		return
	}

	configFile := flag.Arg(0)
	if configFile == "" {
		configFile = "versioninfo.json"
	}
	var input = io.ReadCloser(os.Stdin)
	if configFile != "-" {
		if input, err = os.Open(configFile); err != nil {
			log.Printf("Cannot open %q: %v", configFile, err)
			os.Exit(1)
		}
	}

	// Read the config file.
	jsonBytes, err := ioutil.ReadAll(input)
	input.Close()
	if err != nil {
		log.Printf("Error reading %q: %v", configFile, err)
		os.Exit(1)
	}

	// Create a new container.
	vi := &goversioninfo.VersionInfo{}

	// Parse the config.
	if err := vi.ParseJSON(jsonBytes); err != nil {
		log.Printf("Could not parse the .json file: %v", err)
		os.Exit(2)
	}

	// Override from flags.
	if *flagIcon != "" {
		vi.IconPath = *flagIcon
	}
	if *flagManifest != "" {
		vi.ManifestPath = *flagManifest
	}
	if *flagComment != "" {
		vi.StringFileInfo.Comments = *flagComment
	}
	if *flagCompany != "" {
		vi.StringFileInfo.CompanyName = *flagCompany
	}
	if *flagDescription != "" {
		vi.StringFileInfo.FileDescription = *flagDescription
	}
	if *flagFileVersion != "" {
		vi.StringFileInfo.FileVersion = *flagFileVersion
	}
	if *flagInternalName != "" {
		vi.StringFileInfo.InternalName = *flagInternalName
	}
	if *flagCopyright != "" {
		vi.StringFileInfo.LegalCopyright = *flagCopyright
	}
	if *flagTrademark != "" {
		vi.StringFileInfo.LegalTrademarks = *flagTrademark
	}
	if *flagOriginalName != "" {
		vi.StringFileInfo.OriginalFilename = *flagOriginalName
	}
	if *flagPrivateBuild != "" {
		vi.StringFileInfo.PrivateBuild = *flagPrivateBuild
	}
	if *flagProductName != "" {
		vi.StringFileInfo.ProductName = *flagProductName
	}
	if *flagProductVersion != "" {
		vi.StringFileInfo.ProductVersion = *flagProductVersion
	}
	if *flagSpecialBuild != "" {
		vi.StringFileInfo.SpecialBuild = *flagSpecialBuild
	}

	if *flagTranslation > 0 {
		vi.VarFileInfo.Translation.LangID = goversioninfo.LangID(*flagTranslation)
	}
	if *flagCharset > 0 {
		vi.VarFileInfo.Translation.CharsetID = goversioninfo.CharsetID(*flagCharset)
	}

	// File Version flags.
	if *flagVerMajor >= 0 {
		vi.FixedFileInfo.FileVersion.Major = *flagVerMajor
	}
	if *flagVerMinor >= 0 {
		vi.FixedFileInfo.FileVersion.Minor = *flagVerMinor
	}
	if *flagVerPatch >= 0 {
		vi.FixedFileInfo.FileVersion.Patch = *flagVerPatch
	}
	if *flagVerBuild >= 0 {
		vi.FixedFileInfo.FileVersion.Build = *flagVerBuild
	}

	// Product Version flags.
	if *flagProductVerMajor >= 0 {
		vi.FixedFileInfo.ProductVersion.Major = *flagProductVerMajor
	}
	if *flagProductVerMinor >= 0 {
		vi.FixedFileInfo.ProductVersion.Minor = *flagProductVerMinor
	}
	if *flagProductVerPatch >= 0 {
		vi.FixedFileInfo.ProductVersion.Patch = *flagProductVerPatch
	}
	if *flagProductVerBuild >= 0 {
		vi.FixedFileInfo.ProductVersion.Build = *flagProductVerBuild
	}

	// Fill the structures with config data.
	vi.Build()

	// Write the data to a buffer.
	vi.Walk()

	// List of the architectures to output.
	var archs []string

	// If platform specific, then output all the architectures for Windows.
	if flagPlatformSpecific != nil && *flagPlatformSpecific {
		archs = []string{"386", "amd64"}
	} else {
		// Set the architecture, defaulted to 32-bit.
		archs = []string{"386"} // 32-bit
		if flag64 != nil && *flag64 {
			archs = []string{"amd64"} // 64-bit
		}
	}

	// Loop through each artchitecture.
	for _, item := range archs {
		// Create the file using the -o argument.
		fileout := *flagOut

		// If the number of architectures is greater than one, then don't use
		// the -o argument.
		if len(archs) > 1 {
			fileout = fmt.Sprintf("resource_windows_%v.syso", item)
		}

		// Create the file for the specified architecture.
		if err := vi.WriteSyso(fileout, item); err != nil {
			log.Printf("Error writing syso: %v", err)
			os.Exit(3)
		}
	}
}

const example = `{
	"FixedFileInfo": {
		"FileVersion": {
			"Major": 6,
			"Minor": 3,
			"Patch": 9600,
			"Build": 17284
		},
		"ProductVersion": {
			"Major": 6,
			"Minor": 3,
			"Patch": 9600,
			"Build": 17284
		},
		"FileFlagsMask": "3f",
		"FileFlags ": "00",
		"FileOS": "040004",
		"FileType": "01",
		"FileSubType": "00"
	},
	"StringFileInfo": {
		"Comments": "",
		"CompanyName": "Company, Inc.",
		"FileDescription": "",
		"FileVersion": "6.3.9600.17284 (aaa.140822-1915)",
		"InternalName": "goversioninfo",
		"LegalCopyright": "© Author. Licensed under MIT.",
		"LegalTrademarks": "",
		"OriginalFilename": "goversioninfo",
		"PrivateBuild": "",
		"ProductName": "Go Version Info",
		"ProductVersion": "6.3.9600.17284",
		"SpecialBuild": ""
	},
	"VarFileInfo": {
		"Translation": {
			"LangID": "0409",
			"CharsetID": "04B0"
		}
	}
}`
