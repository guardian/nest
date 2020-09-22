#!/usr/bin/env node

import { CfnOutput, CfnParameter, Construct, Duration, Stack, StackProps } from '@aws-cdk/core';
import { Repository } from '@aws-cdk/aws-ecr';
import { Schedule } from '@aws-cdk/aws-applicationautoscaling';
import { Cluster, Compatibility, ContainerImage, LogDrivers, TaskDefinition } from '@aws-cdk/aws-ecs';
import { Vpc } from '@aws-cdk/aws-ec2';
import { RetentionDays } from '@aws-cdk/aws-logs';
import { Alarm, TreatMissingData } from '@aws-cdk/aws-cloudwatch';
import { SnsAction } from '@aws-cdk/aws-cloudwatch-actions';
import { ServiceIntegrationPattern, StateMachine, Task } from '@aws-cdk/aws-stepfunctions';
import { RunEcsFargateTask } from '@aws-cdk/aws-stepfunctions-tasks';
import { Rule } from '@aws-cdk/aws-events';
import { SfnStateMachine } from '@aws-cdk/aws-events-targets';
import { Topic } from '@aws-cdk/aws-sns';
import { EmailSubscription } from '@aws-cdk/aws-sns-subscriptions';
import { Tags } from "./Tags";

export class FargateScheduledTask extends Stack {
    constructor(scope: Construct, id: string, props: StackProps) {
        super(scope, id, props);

        const tags = new Tags(this, "Guardian core tags");

        const currentBuild = new CfnParameter(this, 'BuildId');

        const taskDefinitionCpu = new CfnParameter(this, "TaskDefinitionCPU", {
            type: "String",
            default: "2048",
        });

        const taskDefinitionMemory = new CfnParameter(this, "TaskDefinitionMemory", {
            type: "String",
            default: "4096",
        });

        const vpcId = new CfnParameter(this, "VpcId", {
            type: "AWS::EC2::VPC::Id",
            description: "VPC in which the task will run",
        });

        const privateSubnet = new CfnParameter(this, "PrivateSubnet", {
            type: "AWS::EC2::Subnet::Id",
            description: "A private subnet (from the VPC) in which the task will run",
        });

        const availabilityZone = new CfnParameter(this, "AZ", {
            type: "AWS::EC2::AvailabilityZone::Name",
            description: "The availability zone where the private subnet resides",
        });

        const alertEmail = new CfnParameter(this, "AlertEmail", {
            type: "String",
            description: "Email address to receive alerts if the task fails",
        });

        const repository = Repository.fromRepositoryName(this, 'Repository', `${tags.stack.valueAsString}-${tags.app.valueAsString}`);

        const taskDefinition = new TaskDefinition(this, 'TaskDefinition', {
            compatibility: Compatibility.FARGATE,
            cpu: taskDefinitionCpu.valueAsString,
            memoryMiB: taskDefinitionMemory.valueAsString
        });

        taskDefinition.addContainer('Container', {
            image: ContainerImage.fromEcrRepository(repository, currentBuild.valueAsString),
            logging: LogDrivers.awsLogs({
                streamPrefix: "tags.app.valueAsString",
                logRetention: RetentionDays.TWO_WEEKS
            })
        });

        const vpc = Vpc.fromVpcAttributes(this, "vpc", {
            vpcId: vpcId.valueAsString,
            availabilityZones: [availabilityZone.valueAsString],
            privateSubnetIds: [privateSubnet.valueAsString]
        });

        const cluster = Cluster.fromClusterAttributes(this, 'default', {
            clusterName: 'default',
            vpc,
            securityGroups: []
        });

        const task = new Task(this, 'Task', {
            task: new RunEcsFargateTask({
                integrationPattern: ServiceIntegrationPattern.SYNC,
                cluster,
                taskDefinition: taskDefinition
            }),
            resultPath: 'DISCARD',
            timeout: Duration.minutes(5)
        });

        const stateMachine = new StateMachine(this, 'StateMachine', {
            definition: task
        });

        new Rule(this, 'ScheduleRule', {
            // TODO MRB: refactor this schedule out to a parameter
            schedule: Schedule.cron({ hour: '7', minute: '30' }),
            targets: [
                new SfnStateMachine(stateMachine)
            ]
        });

        const alarmSnsTopic = new Topic(this, 'AlarmSnsTopic');
        alarmSnsTopic.addSubscription(new EmailSubscription(alertEmail.valueAsString));

        const alarms = [
            {
                name: 'ExecutionsFailedAlarm',
                description: `${tags.app.valueAsString} failed`,
                metric: stateMachine.metricFailed({
                    period: Duration.hours(1),
                    statistic: 'sum',
                })
            },
            {
                name: 'TimeoutAlarm',
                description: `${tags.app.valueAsString} timed out`,
                metric: stateMachine.metricTimedOut({
                    period: Duration.hours(1),
                    statistic: 'sum',
                })
            },
        ];

        alarms.forEach((a => {
            const alarm = new Alarm(this, a.name, {
                alarmDescription: a.description,
                actionsEnabled: true,
                metric: a.metric,
                // default for comparisonOperator is GreaterThanOrEqualToThreshold
                threshold: 1,
                evaluationPeriods: 1,
                treatMissingData: TreatMissingData.NOT_BREACHING,
            });
            alarm.addAlarmAction(new SnsAction(alarmSnsTopic));
        }));

        new CfnOutput(this, 'StateMachineArn', {
            value: stateMachine.stateMachineArn
        });
    }

}
