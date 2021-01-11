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

        const instanceClass = new cdk.CfnParameter(this, "InstanceClass", {
            type: "String",
            default: "t3a",
        });

        const instanceSize = new cdk.CfnParameter(this, "InstanceSize", {
            type: "String",
            default: "small",
        });

        const vpcId = new cdk.CfnParameter(this, "VpcId", {
            type: "AWS::EC2::VPC::Id",
            description:
                "VPC in which instances will run. It should have a least one public subnet.",
        });

        const publicSubnets = new cdk.CfnParameter(this, "Subnets", {
            type: "List<AWS::EC2::Subnet::Id>",
            description: "(Public) Subnets where instances will run.",
        });

        const availabilityZones = new cdk.CfnParameter(this, "AZs", {
            type: "List<AWS::EC2::AvailabilityZone::Name>",
            description:
                "List of AZs. Typically we use eu-west-1a, eu-west-1b, and eu-west-1c here for good availability if one has issues.",
        });

        const ami = new cdk.CfnParameter(this, "AMI", {
            type: "AWS::EC2::Image::Id",
            description:
                "AMI ID to be provded by RiffRaff. Must include: docker and also nest-secrets. Our Amazon Linux 2 Docker recipe is recommended here.",
        });

        const s3Bucket = new cdk.CfnParameter(this, "S3Bucket", {
            type: "String",
            description:
                "Name of S3 bucket where artifact found. This should be the same as the 'artifactBucket' set in your 'nest.json' file.",
        });

        const s3Key = new cdk.CfnParameter(this, "S3Key", {
            type: "String",
            description:
                "S3 key where artifact lives. The required format is: '[stack]/[STAGE]/[app]/app.tar.gz'",
        });

        const tag = new cdk.CfnParameter(this, "DockerTag", {
            type: "String",
            description:
                "Once the s3 artifact is docker loaded, this tag is used to determine which container to start. The required format is: '[app]:latest'.",
        });

        const certificateArn = new cdk.CfnParameter(this, "CertificateArn", {
            type: "String",
            description:
                "ARN of certificate used for the ALB. You will need to create this manually (using ACM) unfortunately and also point the corresponding domain at the ALB itself once the stack is created.",
        });

        const minCapacity = new cdk.CfnParameter(this, "MinCapacity", {
            type: "Number",
            description:
                "Min capacity of ASG. Typically, we want at least 3 instances for PROD for availability purposes, but 1 for CODE.",
            default: 1,
        });

        const maxCapacity = new cdk.CfnParameter(this, "MaxCapacity", {
            type: "Number",
            description:
                "Max capacity of ASG (double normal capacity at least to allow for deploys",
            default: 2,
        });

        const rolePolicyARNs = new cdk.CfnParameter(this, "PolicyARNs", {
            type: "CommaDelimitedList",
            description:
                "ARNs for managed policies you want included in instance role (CURRENTLY THIS DOES NOT WORK).",
        });

        const kmsKey = new cdk.CfnParameter(this, "KMSKey", {
            type: "String",
            description: "KMS key used to decrypt parameter store secrets.",
        });

        const targetCPU = new cdk.CfnParameter(this, "TargetCPU", {
            type: "Number",
            description:
                "Target CPU, used for autoscaling. Nb. you may want to set this quite low if using burstable instances such as t3 ones to avoid paying for lots of CPU credits.",
            default: 80,
        });

        const stageMapping = new cdk.CfnMapping(this, "stages", {
            mapping: {
                CODE: { lower: "code" },
                PROD: { lower: "prod" },
                DEV: { lower: "dev" },
            },
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
                        new iam.PolicyStatement({
                            effect: iam.Effect.ALLOW,
                            resources: [
                                `arn:aws:ssm:eu-west-1:${
                                    this.account
                                }:parameter/${
                                    tags.app.valueAsString
                                }/${cdk.Fn.findInMap(
                                    "stages",
                                    tags.stage.valueAsString,
                                    "lower"
                                )}`,
                            ],
                            actions: ["ssm:GetParametersByPath"],
                        }),
                        new iam.PolicyStatement({
                            effect: iam.Effect.ALLOW,
                            resources: [kmsKey.valueAsString],
                            actions: ["kms:Decrypt"],
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
            `nest-secrets --prefix /${
                tags.app.valueAsString
            }/${cdk.Fn.findInMap(
                "stages",
                tags.stage.valueAsString,
                "lower"
            )} > .env`,
            `aws s3 cp s3://${s3Bucket.valueAsString}/${s3Key.valueAsString} app.tar.gz`,
            `docker load < app.tar.gz`,
            `docker run \
            --ulimit nofile=2048:2048 \
            --env-file .env \
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
            minCapacity: minCapacity.valueAsNumber,
        });

        asg.scaleOnCpuUtilization("GTCPU", {
            targetUtilizationPercent: targetCPU.valueAsNumber,
        });

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
            deregistrationDelay: cdk.Duration.seconds(10),
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
