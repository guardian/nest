// Nest provides helpers for building and deploying Guardian services. Riffraff
// docs are really helpful for understanding what is going on here:
// https://riffraff.gutools.co.uk/docs/reference/s3-artifact-layout.md.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
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

var helpText = `nest [http | build | upload | init | help]
http	starts a basic HTTP service (for testing)
build	generate a Riffraff artifact
upload	upload artifact files to Riffraff S3 bucket and build.json to build bucket
init	helper to generate your nest.json config
help	print this help
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(helpText)
		return
	}

	switch os.Args[1] {
	case "http":
		startTestServer()
	case "build":
		c, err := config.ReadConfig()
		check(err, "Unable to read nest.json config.")
		buildArtifact(c)
	case "upload":
		c, err := config.ReadConfig()
		check(err, "Unable to read nest.json config.")
		uploadArtifact(c)
	case "init":
		err := config.InitConfig()
		check(err, "Unable to init nest.json config.")
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
		BuildNumber: env("BUILD_COUNTER", "1"),
		StartTime:   time.Now().Format(time.RFC3339),
		VCSURL:      c.VCSURL,
		Branch:      env("BRANCH_NAME", "main"),
		Revision:    env("BUILD_VCS_NUMBER", ""),
	}, nil
}

func env(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}

	return fallback
}

func uploadArtifact(c config.Config) {
	buildInfo, err := getBuildInfo(c)
	check(err, "Unable to generate Riffraff build.json file.")

	// upload artifact
	prefix := fmt.Sprintf("%s/%s", buildInfo.ProjectName, buildInfo.BuildNumber)
	err = s3.UploadDir("riffraff-artifact", prefix, target)
	check(err, "Unable to upload artifact files.")

	// upload build info (after artifacts to avoid race conditions in RR)
	buildJSON, err := json.Marshal(buildInfo)
	check(err, "Unable to marshal build.json")

	path := fmt.Sprintf("%s/%s/build.json", buildInfo.ProjectName, buildInfo.BuildNumber)
	err = s3.UploadFile("riffraff-builds", path, bytes.NewReader(buildJSON))
	check(err, "Unable to upload Riffraff build.json file.")
}

// https://riffraff.gutools.co.uk/docs/reference/s3-artifact-layout.md
func buildArtifact(c config.Config) {
	if c.DeploymentType != "service-ec2" {
		fmt.Printf("unsupported deployment type: %s\n", c.DeploymentType)
		os.Exit(1)
	}

	// TODO clean ./target dir first

	artifactFile := "app.tar.gz"

	makeDir(target, c.App)
	makeDir(target, "cfn")

	buildOut, err := exec.Command("docker", "build", "-t", fmt.Sprintf("%s:latest", c.App), ".").Output()
	check(err, fmt.Sprintf("Unable to build Docker image: %s.", string(buildOut)))

	saveOut, err := exec.Command("bash", "-c", fmt.Sprintf("docker save %s:latest | gzip > %s", c.App, artifactFile)).Output()
	check(err, fmt.Sprintf("Unable to save Docker image: %s.", string(saveOut)))

	tmpl, _ := template.New("riffraff").Parse(tpl.RiffRaff)

	rr := bytes.Buffer{}
	tmpl.Execute(&rr, info{App: c.App, Bucket: c.ArtifactBucket})
	rrOutput, err := ioutil.ReadAll(&rr)
	check(err, "Unable to read Riffraff template output.")

	err = ioutil.WriteFile(filepath.Join(target, "riff-raff.yaml"), rrOutput, os.ModePerm)
	check(err, "Unable to write riff-raff.yaml file.")

	err = ioutil.WriteFile(filepath.Join(target, "cfn", "cfn.yaml"), []byte(tpl.Cfn), os.ModePerm)
	check(err, "Unable to write cfn.yaml file.")

	err = os.Rename(artifactFile, filepath.Join(target, c.App, artifactFile))
	check(err, "Unable to move artifact.")
}

func makeDir(target, folder string) {
	path := filepath.Join(target, folder)
	err := os.MkdirAll(path, os.ModePerm)
	check(err, fmt.Sprintf("Unable to make directories %s.", path))
}

// TODO add second argument as helper message on failure
func check(err error, msg string) {
	if err != nil {
		fmt.Println(msg)
		fmt.Println(err)
		os.Exit(1)
	}
}

func startTestServer() {
	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world")
	})

	log.Fatal(http.ListenAndServe(":"+env("PORT", "8080"), nil))
}
