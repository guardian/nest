package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config is the type of a Nest config file (typically nest.json).
type Config struct {
	App            string `json:"app"`
	Stack          string `json:"stack"`
	VCSURL         string `json:"vcsURL"`
	DeploymentType string `json:"deploymentType"`
	ArtifactBucket string `json:"artifactBucket"`

	// Note, leave empty to use the default. This is really for migrations only.
	CloudformationStackName string `json:"cloudformationStackName"`
}

// ReadConfig reads a nest.json file (if one exists) in the current directory.
func ReadConfig() (Config, error) {
	var config Config

	bytes, err := ioutil.ReadFile("nest.json")
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(bytes, &config)
	return config, err
}

// InitConfig is a helper to create a nest.json configuration file.
func InitConfig() error {
	var config Config

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter app: ")
	app, _, _ := reader.ReadLine()
	config.App = string(app)

	fmt.Print("Enter stack: ")
	stack, _, _ := reader.ReadLine()
	config.Stack = string(stack)

	fmt.Print("Enter VCS URL: ")
	VCSURL, _, _ := reader.ReadLine()
	config.VCSURL = string(VCSURL)

	fmt.Print("Enter artifact bucket: ")
	bucket, _, _ := reader.ReadLine()
	config.ArtifactBucket = string(bucket)

	config.DeploymentType = "alb-ec2-service"

	data, _ := json.MarshalIndent(config, "", "    ")

	return ioutil.WriteFile("nest.json", data, os.ModePerm)

}
