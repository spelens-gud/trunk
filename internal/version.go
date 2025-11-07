package internal

import (
	"fmt"
	"os"
)

var (
	GoVersion  = ""
	BuildDate  = ""
	GitVersion = ""
)

// go build -ldflags "-X 'internal/version.GoVersion=${goversion}' -X 'internal/version.BuildDate=${gitlasttime}' -X 'internal/version.GitVersion=${gitinfo}'"



func GetVersionInfo() {
	fmt.Println("Go  Version :", GoVersion)
	fmt.Println("Build  Date :", BuildDate)
	fmt.Println("Svn Version :", GitVersion)
	os.Exit(1)
}
