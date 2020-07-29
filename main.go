package main

import (
	"os"
	"text/template"
)

// artifact
//   riff-raff.yaml
//   cfn/cfn.yaml
//   image/image.tar.gz

type info struct {
	App    string
	Bucket string
}

var riffraffTpl string = `
stacks: [frontend]
regions: [eu-west-1]

deployments:
    cfn:
        type: cloud-formation
        app: {{.App}}
        parameters:
            cloudFormationStackName: {{.App}}
            templatePath: cloudformation.yaml
            cloudFormationStackByTags: false
    {{.App}}:
        type: autoscaling
        dependencies: [cfn]
        parameters:
            bucket: {{.Bucket}}

`

func main() {
	tmpl, _ := template.New("riffraff").Parse(riffraffTpl)
	tmpl.Execute(os.Stdout, info{App: "contributions-service", Bucket: "aws-frontend-artifacts"})
}
