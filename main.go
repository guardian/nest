package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/guardian/nest/s3"
	"github.com/guardian/nest/tpl"
)

type info struct {
	App    string
	Bucket string
}

var target string = "target"

func main() {
	switch os.Args[1] {
	case "build":
		c := readConfig()
		buildArtifact(c)
	case "upload":
		c := readConfig()
		uploadArtifact(c)
	case "init":
		fmt.Println("Not implemented yet!")
		os.Exit(1)
	}
}

// Config is the type of a Nest config file (typically nest.json)
type Config struct {
	App          string `json:"app"`
	Stack        string `json:"stack"`
	ArtifactPath string `json:"artifactPath"`
	VCSURL       string `json:"vcsURL"`
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

func readConfig() Config {
	bytes, err := ioutil.ReadFile("nest.json")
	check(err)

	var config Config
	err = json.Unmarshal(bytes, &config)
	check(err)

	return config
}

// TODO only works on TC probably atm
func getBuildInfo(c Config) (BuildInfo, error) {
	return BuildInfo{
		ProjectName: fmt.Sprintf("%s:%s", c.Stack, c.App),
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

// https://riffraff.gutools.co.uk/docs/reference/s3-artifact-layout.md
func uploadArtifact(c Config) {
	// upload artifact
	err := s3.UploadDir("riffraff-artifacts", "prefix", target)
	check(err)

	// upload build.json
	buildInfo, err := getBuildInfo(c)
	check(err)

	buildJSON, _ := json.Marshal(buildInfo)
	path := fmt.Sprintf("%s/%s/build.json", buildInfo.ProjectName, buildInfo.BuildNumber)
	err = s3.UploadFile("riffraff-builds", path, bytes.NewReader(buildJSON))
	check(err)
}

// https://riffraff.gutools.co.uk/docs/reference/s3-artifact-layout.md
func buildArtifact(c Config) {
	makeDir(target, c.App)
	makeDir(target, "cfn")

	tmpl, _ := template.New("riffraff").Parse(tpl.RiffRaff)

	rr := bytes.Buffer{}
	tmpl.Execute(&rr, info{App: "contributions-service", Bucket: "aws-frontend-artifacts"})
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
		log.Fatal(err)
	}
}
