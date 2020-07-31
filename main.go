// Nest provides helpers for building and deploying Guardian services. Riffraff
// docs are really helpful for understanding what is going on here:
// https://riffraff.gutools.co.uk/docs/reference/s3-artifact-layout.md.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/guardian/nest/config"
	"github.com/guardian/nest/s3"
	"github.com/guardian/nest/tpl"
)

type info struct {
	App    string
	Bucket string
}

var target string = "target"

var helpText = `nest [build | upload | init | help]
build	generate a Riffraff artifact
upload	upload artifact files to Riffraff S3 bucket and build.json to build bucket
init	(TODO) helper to generate your nest.json config
help	print this help
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(helpText)
		return
	}

	switch os.Args[1] {
	case "build":
		c, err := config.ReadConfig()
		check(err)
		buildArtifact(c)
	case "upload":
		c, err := config.ReadConfig()
		check(err)
		uploadArtifact(c)
	case "init":
		err := config.InitConfig()
		check(err)
	case "help":
		fmt.Print(helpText)
	}
}

// BuildInfo - see https://riffraff.gutools.co.uk/docs/reference/build.json.md
type BuildInfo struct {
	ProjectName string `json:"projectName"`
	BuildNumber string `json:"buildNumber"`
	StartTime   string `json:"startTime"`
	VCSURL      string `json:"vcsURL"`
	Branch      string `json:"branch"`
	Revision    string `json:"revision"`
}

// TODO only works on TC probably atm
func getBuildInfo(c config.Config) (BuildInfo, error) {
	return BuildInfo{
		ProjectName: fmt.Sprintf("%s::%s", c.Stack, c.App),
		BuildNumber: env("BUILD_VCS_NUMBER", "1"),
		StartTime:   time.Now().Format(time.RFC3339),
		VCSURL:      c.VCSURL,
		Branch:      env("BRANCH_NAME", "main"),
		Revision:    env("BUILD_VCS_NUMBER", ""),
	}, nil
}

func env(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return fallback
}

func uploadArtifact(c config.Config) {
	buildInfo, err := getBuildInfo(c)
	check(err)

	// upload artifact
	prefix := fmt.Sprintf("%s/%s", buildInfo.ProjectName, buildInfo.BuildNumber)
	err = s3.UploadDir("riffraff-artifacts", prefix, target)
	check(err)

	// upload build info (after artifacts to avoid race conditions in RR)
	buildJSON, _ := json.Marshal(buildInfo)
	path := fmt.Sprintf("%s/%s/build.json", buildInfo.ProjectName, buildInfo.BuildNumber)
	err = s3.UploadFile("riffraff-builds", path, bytes.NewReader(buildJSON), true)
	check(err)
}

// https://riffraff.gutools.co.uk/docs/reference/s3-artifact-layout.md
func buildArtifact(c config.Config) {
	if c.DeploymentType != "service-ec2" {
		fmt.Printf("unsupported deployment type: %s\n", c.DeploymentType)
		os.Exit(1)
	}

	makeDir(target, c.App)
	makeDir(target, "cfn")

	tmpl, _ := template.New("riffraff").Parse(tpl.RiffRaff)

	rr := bytes.Buffer{}
	tmpl.Execute(&rr, info{App: c.App, Bucket: c.ArtifactBucket})
	rrOutput, err := ioutil.ReadAll(&rr)
	check(err)

	err = ioutil.WriteFile(filepath.Join(target, "riff-raff.yaml"), rrOutput, os.ModePerm)
	check(err)

	err = ioutil.WriteFile(filepath.Join(target, "cfn", "cfn.yaml"), []byte(tpl.Cfn), os.ModePerm)
	check(err)

	err = os.Rename(c.ArtifactPath, filepath.Join(target, c.App, c.ArtifactPath))
	check(err)
}

func makeDir(target, folder string) {
	err := os.MkdirAll(filepath.Join(target, folder), os.ModePerm)
	check(err)
}

// TODO add second argument as helper message on failure
func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
