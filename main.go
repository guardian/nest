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
	App                     string
	Bucket                  string
	CloudformationStackName string
	Stack                   string
}

var target string = "target"

var helpText = `nest [http | build | upload | init | recipes | help]
http	starts a basic HTTP service (for testing)
build	generate a Riffraff artifact
upload	upload artifact files to Riffraff S3 bucket and build.json to build bucket
init	helper to generate your nest.json config
recipes	describes supported deployment types
help	print this help
`

var recipesText = `nest recipes are predefined deployment types:

alb-ec2-service
  Suitable for a service that is exposed over HTTP. It expects a 
  Dockerfile in your root directory that starts your service and
  outputs a deployable Riffraff artifact that creates an ASG with
  your app inside. An ALB is used to front things.

  Logging is provided via Cloudwatch Logs (container output is
  redirected here). You may want to forward this on to the Guardian
  ELK stack from Cloudwatch.

  At the moment, you will need to manually create the Cloudformation
  stack (use 'nest build' to generate) before deploying via Riffraff
  in order to specify the required parameter values - i.e. for things
  like the VPC ID and subnets.

  Use the naming convention: <STACK>-<APP>-<STAGE>
  (e.g. frontend-nest-PROD).

fargate-scheduled-task
  Run a task on a schedule. It expects a Dockerfile in your root
  directory that runs the task. Riff-Raff integration will deploy the
  entire generated Cloudformation template, setting the BuildId parameter
  to the version deployed.

  Logging is provided via Cloudwatch Logs (container output is
  redirected here). You may want to forward this on to the Guardian
  ELK stack from Cloudwatch.

  At the moment there are a couple of manual steps:
	- Create an ECR repository named <STACK>-<APP> (eg frontend-nest)
	- Manually deploy the Cloudformation stack before deploying via Riff-Raff
	  - Use 'nest build' to generate it
	  - This is required to specify parameter values (VPC ID, subnets)
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
	case "recipes":
		fmt.Print(recipesText)
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
		StartTime:   time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
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
func buildAlbEc2ServiceArtifact(c config.Config) {
	// TODO clean ./target dir first

	artifactFile := "app.tar.gz"

	makeDir(target, c.App)
	makeDir(target, "cfn")

	buildOut, err := exec.Command("docker", "build", "-t", fmt.Sprintf("%s:latest", c.App), ".").Output()
	check(err, fmt.Sprintf("Unable to build Docker image: %s.", string(buildOut)))

	saveOut, err := exec.Command("bash", "-c", fmt.Sprintf("docker save %s:latest | gzip > %s", c.App, artifactFile)).Output()
	check(err, fmt.Sprintf("Unable to save Docker image: %s.", string(saveOut)))

	tmpl, _ := template.New("riffraff").Parse(tpl.RiffRaffAlbEc2Service)

	rr := bytes.Buffer{}
	cfnStackName := c.CloudformationStackName
	if cfnStackName == "" {
		cfnStackName = c.App
	}
	tmpl.Execute(&rr, info{App: c.App, Bucket: c.ArtifactBucket, CloudformationStackName: cfnStackName, Stack: c.Stack})
	rrOutput, err := ioutil.ReadAll(&rr)
	check(err, "Unable to read Riffraff template output.")

	err = ioutil.WriteFile(filepath.Join(target, "riff-raff.yaml"), rrOutput, os.ModePerm)
	check(err, "Unable to write riff-raff.yaml file.")

	err = ioutil.WriteFile(filepath.Join(target, "cfn", "cfn.yaml"), []byte(tpl.AlbEc2Stack), os.ModePerm)
	check(err, "Unable to write cfn.yaml file.")

	err = os.Rename(artifactFile, filepath.Join(target, c.App, artifactFile))
	check(err, "Unable to move artifact.")
}

func buildFargateScheduledTaskArtifact(c config.Config) {
	buildNumber := "unknown"
	buildInfo, err := getBuildInfo(c)
	if err == nil {
		buildNumber = buildInfo.BuildNumber
	}

	makeDir(target, c.App)
	makeDir(target, "cfn")

	containerTag := fmt.Sprintf("%s.dkr.ecr.eu-west-1/%s:%s", c.AccountId, c.App, buildNumber)

	buildOut, err := exec.Command("docker", "build", "-t", containerTag, ".").Output()
	check(err, fmt.Sprintf("Unable to build Docker image: %s.", string(buildOut)))

	tmpl, _ := template.New("riffraff").Parse(tpl.RiffRaffFargateScheduledTask)

	rr := bytes.Buffer{}
	tmpl.Execute(&rr, info{App: c.App, Stack: c.Stack})
	rrOutput, err := ioutil.ReadAll(&rr)
	check(err, "Unable to read Riffraff template output.")

	err = ioutil.WriteFile(filepath.Join(target, "riff-raff.yaml"), rrOutput, os.ModePerm)
	check(err, "Unable to write riff-raff.yaml file.")

	err = ioutil.WriteFile(filepath.Join(target, "cfn", "cfn.yaml"), []byte(tpl.FargateScheduledTask), os.ModePerm)
	check(err, "Unable to write cfn.yaml file.")
}

func buildArtifact(c config.Config) {
	switch c.DeploymentType {
		case "alb-ec2-service":
			buildAlbEc2ServiceArtifact(c)

		case "fargate-scheduled-task":
			buildFargateScheduledTaskArtifact(c)

		default:
			fmt.Printf("unsupported deployment type: %s\n", c.DeploymentType)
			os.Exit(1)
	}
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
		name := env("nest_prod_name", "Fitzchivalry")
		fmt.Fprintf(w, "Hello, %s", name)
	})

	log.Fatal(http.ListenAndServe(":"+env("PORT", "3030"), nil))
}
