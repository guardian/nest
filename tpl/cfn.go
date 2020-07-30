package tpl

var Cfn string = `
{
	"Parameters": {
	  "Stack": {
		"Type": "String",
		"Default": "frontend"
	  },
	  "Stage": {
		"Type": "String",
		"Default": "PROD"
	  },
	  "App": {
		"Type": "String"
	  },
	  "Instanceclass": {
		"Type": "String",
		"Default": "t3a"
	  },
	  "Instancesize": {
		"Type": "String",
		"Default": "small"
	  },
	  "VpcId": {
		"Type": "AWS::EC2::VPC::Id",
		"Description": "VPC in which instances will run"
	  },
	  "Subnets": {
		"Type": "List<AWS::EC2::Subnet::Id>",
		"Description": "Subnets where instances will run"
	  },
	  "AZs": {
		"Type": "List<AWS::EC2::AvailabilityZone::Name>",
		"Description": "List of AZs"
	  },
	  "AMI": {
		"Type": "AWS::EC2::Image::Id",
		"Default": "ami-07e05aef825d2078a",
		"Description": "AMI ID to be provded by RiffRaff. Should include Docker at least."
	  },
	  "S3Bucket": {
		"Type": "String",
		"Description": "Name of S3 bucket where artifact found"
	  },
	  "S3Key": {
		"Type": "String",
		"Description": "S3 key where artifact lives (should be a Docker saved .tar file)"
	  },
	  "DockerTag": {
		"Type": "String",
		"Description": "Once the s3 artifact is docker loaded, this tag is used to determine which container to start"
	  },
	  "PolicyARNs": {
		"Type": "CommaDelimitedList",
		"Description": "ARNs for managed policies you want included in instance role"
	  }
	},
	"Resources": {
	  "roleC7B7E775": {
		"Type": "AWS::IAM::Role",
		"Properties": {
		  "AssumeRolePolicyDocument": {
			"Statement": [
			  {
				"Action": "sts:AssumeRole",
				"Effect": "Allow",
				"Principal": {
				  "Service": "ec2.amazonaws.com"
				}
			  }
			],
			"Version": "2012-10-17"
		  },
		  "Policies": [
			{
			  "PolicyDocument": {
				"Statement": [
				  {
					"Action": [
					  "s3:Get*",
					  "s3:List*",
					  "s3:HeadObject"
					],
					"Effect": "Allow",
					"Resource": [
					  {
						"Fn::Join": [
						  "",
						  [
							"arn:aws:s3:::",
							{
							  "Ref": "S3Bucket"
							}
						  ]
						]
					  },
					  {
						"Fn::Join": [
						  "",
						  [
							"arn:aws:s3:::",
							{
							  "Ref": "S3Bucket"
							},
							"/*"
						  ]
						]
					  }
					]
				  },
				  {
					"Action": [
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
					  "ssmmessages:OpenDataChannel"
					],
					"Effect": "Allow",
					"Resource": "*"
				  },
				  {
					"Action": [
					  "logs:CreateLogGroup",
					  "logs:CreateLogStream",
					  "logs:PutLogEvents"
					],
					"Effect": "Allow",
					"Resource": "*"
				  }
				],
				"Version": "2012-10-17"
			  },
			  "PolicyName": "required"
			}
		  ],
		  "Tags": [
			{
			  "Key": "App",
			  "Value": {
				"Ref": "App"
			  }
			},
			{
			  "Key": "Stack",
			  "Value": {
				"Ref": "Stack"
			  }
			},
			{
			  "Key": "Stage",
			  "Value": {
				"Ref": "Stage"
			  }
			}
		  ]
		},
		"Metadata": {
		  "aws:cdk:path": "HTTPService/role/Resource"
		}
	  },
	  "ASGInstanceSecurityGroup0525485D": {
		"Type": "AWS::EC2::SecurityGroup",
		"Properties": {
		  "GroupDescription": "HTTPService/ASG/InstanceSecurityGroup",
		  "SecurityGroupEgress": [
			{
			  "CidrIp": "0.0.0.0/0",
			  "Description": "Allow all outbound traffic by default",
			  "IpProtocol": "-1"
			}
		  ],
		  "Tags": [
			{
			  "Key": "App",
			  "Value": {
				"Ref": "App"
			  }
			},
			{
			  "Key": "Name",
			  "Value": "HTTPService/ASG"
			},
			{
			  "Key": "Stack",
			  "Value": {
				"Ref": "Stack"
			  }
			},
			{
			  "Key": "Stage",
			  "Value": {
				"Ref": "Stage"
			  }
			}
		  ],
		  "VpcId": {
			"Ref": "VpcId"
		  }
		},
		"Metadata": {
		  "aws:cdk:path": "HTTPService/ASG/InstanceSecurityGroup/Resource"
		}
	  },
	  "ASGInstanceProfile0A2834D7": {
		"Type": "AWS::IAM::InstanceProfile",
		"Properties": {
		  "Roles": [
			{
			  "Ref": "roleC7B7E775"
			}
		  ]
		},
		"Metadata": {
		  "aws:cdk:path": "HTTPService/ASG/InstanceProfile"
		}
	  },
	  "ASGLaunchConfigC00AF12B": {
		"Type": "AWS::AutoScaling::LaunchConfiguration",
		"Properties": {
		  "ImageId": {
			"Ref": "AMI"
		  },
		  "InstanceType": {
			"Fn::Join": [
			  "",
			  [
				{
				  "Ref": "Instanceclass"
				},
				".",
				{
				  "Ref": "Instancesize"
				}
			  ]
			]
		  },
		  "AssociatePublicIpAddress": true,
		  "IamInstanceProfile": {
			"Ref": "ASGInstanceProfile0A2834D7"
		  },
		  "SecurityGroups": [
			{
			  "Fn::GetAtt": [
				"ASGInstanceSecurityGroup0525485D",
				"GroupId"
			  ]
			}
		  ],
		  "UserData": {
			"Fn::Base64": {
			  "Fn::Join": [
				"",
				[
				  "#!/bin/bash\naws s3 cp s3://",
				  {
					"Ref": "S3Bucket"
				  },
				  "/",
				  {
					"Ref": "S3Key"
				  },
				  " app.tar.gz\ndocker load < app.tar.gz\ndocker run       -p 3030:3030       --log-driver=awslogs       --log-opt awslogs-group=",
				  {
					"Ref": "Stack"
				  },
				  "/",
				  {
					"Ref": "App"
				  },
				  "/",
				  {
					"Ref": "Stage"
				  },
				  "       --log-opt awslogs-create-group=true       ",
				  {
					"Ref": "DockerTag"
				  }
				]
			  ]
			}
		  }
		},
		"DependsOn": [
		  "roleC7B7E775"
		],
		"Metadata": {
		  "aws:cdk:path": "HTTPService/ASG/LaunchConfig"
		}
	  },
	  "ASG46ED3070": {
		"Type": "AWS::AutoScaling::AutoScalingGroup",
		"Properties": {
		  "MaxSize": "2",
		  "MinSize": "1",
		  "LaunchConfigurationName": {
			"Ref": "ASGLaunchConfigC00AF12B"
		  },
		  "Tags": [
			{
			  "Key": "App",
			  "PropagateAtLaunch": true,
			  "Value": {
				"Ref": "App"
			  }
			},
			{
			  "Key": "Name",
			  "PropagateAtLaunch": true,
			  "Value": "HTTPService/ASG"
			},
			{
			  "Key": "Stack",
			  "PropagateAtLaunch": true,
			  "Value": {
				"Ref": "Stack"
			  }
			},
			{
			  "Key": "Stage",
			  "PropagateAtLaunch": true,
			  "Value": {
				"Ref": "Stage"
			  }
			}
		  ],
		  "VPCZoneIdentifier": {
			"Ref": "Subnets"
		  }
		},
		"UpdatePolicy": {
		  "AutoScalingScheduledAction": {
			"IgnoreUnmodifiedGroupSizeProperties": true
		  }
		},
		"Metadata": {
		  "aws:cdk:path": "HTTPService/ASG/ASG"
		}
	  },
	  "ASGScalingPolicyGT80CPUD8CC7169": {
		"Type": "AWS::AutoScaling::ScalingPolicy",
		"Properties": {
		  "AutoScalingGroupName": {
			"Ref": "ASG46ED3070"
		  },
		  "PolicyType": "TargetTrackingScaling",
		  "TargetTrackingConfiguration": {
			"PredefinedMetricSpecification": {
			  "PredefinedMetricType": "ASGAverageCPUUtilization"
			},
			"TargetValue": 80
		  }
		},
		"Metadata": {
		  "aws:cdk:path": "HTTPService/ASG/ScalingPolicyGT80CPU/Resource"
		}
	  }
	}
  }
`
