package tpl

// RiffRaff config template
var RiffRaff string = `
stacks: [{{.Stack}}]
regions: [eu-west-1]

deployments:
    cfn:
        type: cloud-formation
        app: {{.App}}
        parameters:
            cloudFormationStackName: {{.CloudformationStackName}}
            templatePath: cfn.yaml
            cloudFormationStackByTags: false
            amiTags:
                Recipe: amazon-linux-2-x86-docker
                AmigoStage: PROD
    {{.App}}:
        type: autoscaling
        dependencies: [cfn]
        parameters:
            bucket: {{.Bucket}}

`
