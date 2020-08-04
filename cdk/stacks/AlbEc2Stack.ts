import * as cdk from "@aws-cdk/core";
import * as ec2 from "@aws-cdk/aws-ec2";
import * as autoscaling from "@aws-cdk/aws-autoscaling";
import * as iam from "@aws-cdk/aws-iam";
import * as elbv2 from "@aws-cdk/aws-elasticloadbalancingv2";
import { Tags } from "./Tags";

export class AlbEc2Stack extends cdk.Stack {
    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const tags = new Tags(this, "Guardian core tags");

        const instanceClass = new cdk.CfnParameter(this, "Instance class", {
            type: "String",
            default: "t3a",
        });

        const instanceSize = new cdk.CfnParameter(this, "Instance size", {
            type: "String",
            default: "small",
        });

        const vpcId = new cdk.CfnParameter(this, "VpcId", {
            type: "AWS::EC2::VPC::Id",
            description: "VPC in which instances will run",
        });

        const publicSubnets = new cdk.CfnParameter(this, "Subnets", {
            type: "List<AWS::EC2::Subnet::Id>",
            description: "Subnets where instances will run",
        });

        const availabilityZones = new cdk.CfnParameter(this, "AZs", {
            type: "List<AWS::EC2::AvailabilityZone::Name>",
            description: "List of AZs",
        });

        const ami = new cdk.CfnParameter(this, "AMI", {
            type: "AWS::EC2::Image::Id",
            description:
                "AMI ID to be provded by RiffRaff. Should include Docker at least. Our Amazon Linux 2 Docker recipe is recommended here.",
        });

        const s3Bucket = new cdk.CfnParameter(this, "S3 Bucket", {
            type: "String",
            description: "Name of S3 bucket where artifact found",
        });

        const s3Key = new cdk.CfnParameter(this, "S3 Key", {
            type: "String",
            description:
                "S3 key where artifact lives (should be a Docker saved .tar file)",
        });

        const tag = new cdk.CfnParameter(this, "Docker Tag", {
            type: "String",
            description:
                "Once the s3 artifact is docker loaded, this tag is used to determine which container to start",
        });

        const certificateArn = new cdk.CfnParameter(this, "CertificateArn", {
            type: "String",
        });

        const maxCapacity = new cdk.CfnParameter(this, "ASG max capacity", {
            type: "Number",
            description:
                "Max capacity of ASG (double normal capacity at least to allow for deploys",
            default: 2,
        });

        const rolePolicyARNs = new cdk.CfnParameter(this, "Policy ARNs", {
            type: "CommaDelimitedList",
            description:
                "ARNs for managed policies you want included in instance role",
        });

        /*     const managedPolicies = rolePolicyARNs.valueAsList.map((arn) => ({
      managedPolicyArn: arn,
    })); */

        const role = new iam.Role(this, "role", {
            assumedBy: new iam.ServicePrincipal("ec2.amazonaws.com"),
            //managedPolicies: managedPolicies,
            inlinePolicies: {
                required: new iam.PolicyDocument({
                    statements: [
                        new iam.PolicyStatement({
                            effect: iam.Effect.ALLOW,
                            resources: [
                                `arn:aws:s3:::${s3Bucket.valueAsString}`,
                                `arn:aws:s3:::${s3Bucket.valueAsString}/*`,
                            ],
                            actions: ["s3:Get*", "s3:List*", "s3:HeadObject"],
                        }),
                        new iam.PolicyStatement({
                            effect: iam.Effect.ALLOW,
                            resources: ["*"],
                            actions: [
                                "ec2messages:AcknowledgeMessage",
                                "ec2messages:DeleteMessage",
                                "ec2messages:FailMessage",
                                "ec2messages:GetEndpoint",
                                "ec2messages:GetMessages",
                                "ec2messages:SendReply",
                                "ssm:UpdateInstanceInformation",
                                "ssm:ListInstanceAssociations",
                                "ssm:DescribeInstanceProperties",
                                "ssm:DescribeDocumentParameters",
                                "ssmmessages:CreateControlChannel",
                                "ssmmessages:CreateDataChannel",
                                "ssmmessages:OpenControlChannel",
                                "ssmmessages:OpenDataChannel",
                            ],
                        }),
                        new iam.PolicyStatement({
                            effect: iam.Effect.ALLOW,
                            resources: ["*"],
                            actions: [
                                "logs:CreateLogGroup",
                                "logs:CreateLogStream",
                                "logs:PutLogEvents",
                            ],
                        }),
                    ],
                }),
            },
        });

        const vpc = ec2.Vpc.fromVpcAttributes(this, "vpc", {
            vpcId: vpcId.valueAsString,
            availabilityZones: availabilityZones.valueAsList,
            publicSubnetIds: publicSubnets.valueAsList,
        });

        const userData = ec2.UserData.forLinux();
        userData.addCommands(
            `aws s3 cp s3://${s3Bucket.valueAsString}/${s3Key.valueAsString} app.tar.gz`,
            `docker load < app.tar.gz`,
            `docker run \
            -p 3030:3030 \
            --log-driver=awslogs \
            --log-opt awslogs-group=${tags.stack.valueAsString}/${tags.app.valueAsString}/${tags.stage.valueAsString} \
            --log-opt awslogs-create-group=true \
            ${tag.valueAsString}`
        );

        const asg = new autoscaling.AutoScalingGroup(this, "ASG", {
            vpc,
            instanceType: ec2.InstanceType.of(
                instanceClass.valueAsString as ec2.InstanceClass,
                instanceSize.valueAsString as ec2.InstanceSize
            ),
            machineImage: ec2.MachineImage.genericLinux({
                "eu-west-1": ami.valueAsString,
            }),
            userData: userData,
            role: role,
            vpcSubnets: { subnetType: ec2.SubnetType.PUBLIC },
            associatePublicIpAddress: true,
            maxCapacity: maxCapacity.valueAsNumber,
        });

        asg.scaleOnCpuUtilization("GT80CPU", { targetUtilizationPercent: 80 });

        const lb = new elbv2.ApplicationLoadBalancer(this, "LB", {
            vpc,
            internetFacing: true,
            loadBalancerName: `${tags.app.valueAsString}-${tags.stage.valueAsString}`,
        });

        const listener = lb.addListener("Listener", {
            port: 443,
            certificateArns: [certificateArn.value.toString()],
        });

        listener.addTargets("Target", {
            port: 3030,
            protocol: elbv2.ApplicationProtocol.HTTP,
            targets: [asg],
            healthCheck: {
                path: "/healthcheck",
                healthyThresholdCount: 2,
                unhealthyThresholdCount: 5,
                interval: cdk.Duration.seconds(30),
                timeout: cdk.Duration.seconds(10),
            },
        });
    }
}
