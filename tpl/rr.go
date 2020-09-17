package tpl

// RiffRaff config template
var RiffRaffAlbEc2Service string = `
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

var RiffRaffFargateScheduledTask string = `
stacks: [{{.Stack}}]
regions: [eu-west-1]

deployments:
    cfn:
        type: cloud-formation
        app: {{.App}}
        parameters:
            templatePath: cfn.yaml
            createStackIfAbsent: false

`