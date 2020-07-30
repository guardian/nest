package tpl

var RiffRaff string = `
stacks: [frontend]
regions: [eu-west-1]

deployments:
    cfn:
        type: cloud-formation
        app: {{.App}}
        parameters:
            cloudFormationStackName: {{.App}}
            templatePath: cfn.yaml
            cloudFormationStackByTags: false
    {{.App}}:
        type: autoscaling
        dependencies: [cfn]
        parameters:
            bucket: {{.Bucket}}

`
