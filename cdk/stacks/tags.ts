import * as cdk from "@aws-cdk/core";

// Tags applies standard Guardian tags to everything
export class Tags {
    public app: cdk.CfnParameter;
    public stack: cdk.CfnParameter;
    public stage: cdk.CfnParameter;

    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
        const stack = new cdk.CfnParameter(scope, "Stack", {
            type: "String",
            default: "frontend",
        });

        const stage = new cdk.CfnParameter(scope, "Stage", {
            type: "String",
            default: "PROD",
        });

        const app = new cdk.CfnParameter(scope, "App", {
            type: "String",
        });

        this.app = app;
        this.stack = stack;
        this.stage = stage;

        const tags = [
            { key: "Stack", value: stack.valueAsString },
            { key: "Stage", value: stage.valueAsString },
            { key: "App", value: app.valueAsString },
        ];

        tags.forEach((tag) => cdk.Tag.add(scope, tag.key, tag.value));
    }
}
