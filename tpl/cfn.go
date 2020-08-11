package tpl

// AlbEc2Stack - generated from cdk directory
var AlbEc2Stack string = `
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
	  "InstanceClass": {
		"Type": "String",
		"Default": "t3a"
	  },
	  "InstanceSize": {
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
		"Description": "AMI ID to be provded by RiffRaff. Should include Docker at least. Our Amazon Linux 2 Docker recipe is recommended here."
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
	  "CertificateArn": {
		"Type": "String"
	  },
	  "MinCapacity": {
		"Type": "Number",
		"Default": 1,
		"Description": "Min capacity of ASG. Typically, we want at least 3 instances for PROD for availability purposes, but 1 for CODE."
	  },
	  "MaxCapacity": {
		"Type": "Number",
		"Default": 2,
		"Description": "Max capacity of ASG (double normal capacity at least to allow for deploys"
	  },
	  "PolicyARNs": {
		"Type": "CommaDelimitedList",
		"Description": "ARNs for managed policies you want included in instance role"
	  },
	  "KMSKey": {
		"Type": "String",
		"Description": "KMS key used to decrypt parameter store secrets"
	  },
	  "TargetCPU": {
		"Type": "Number",
		"Default": 80,
		"Description": "Target CPU, used for autoscaling. Nb. you may want to set this quite low if using Burstable instances such as t3 ones."
	  }
	},
	"Mappings": {
	  "stages": {
		"CODE": {
		  "lower": "code"
		},
		"PROD": {
		  "lower": "prod"
		},
		"DEV": {
		  "lower": "dev"
		}
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
				  },
				  {
					"Action": "ssm:GetParametersByPath",
					"Effect": "Allow",
					"Resource": {
					  "Fn::Join": [
						"",
						[
						  "arn:aws:ssm:eu-west-1:642631414762:parameter/",
						  {
							"Ref": "App"
						  },
						  "/",
						  {
							"Fn::FindInMap": [
							  "stages",
							  {
								"Ref": "Stage"
							  },
							  "lower"
							]
						  }
						]
					  ]
					}
				  },
				  {
					"Action": "kms:Decrypt",
					"Effect": "Allow",
					"Resource": {
					  "Ref": "KMSKey"
					}
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
		  "aws:cdk:path": "AlbEc2Stack/role/Resource"
		}
	  },
	  "ASGInstanceSecurityGroup0525485D": {
		"Type": "AWS::EC2::SecurityGroup",
		"Properties": {
		  "GroupDescription": "AlbEc2Stack/ASG/InstanceSecurityGroup",
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
			  "Value": "AlbEc2Stack/ASG"
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
		  "aws:cdk:path": "AlbEc2Stack/ASG/InstanceSecurityGroup/Resource"
		}
	  },
	  "ASGInstanceSecurityGroupfromAlbEc2StackLBSecurityGroup7075EADF3030EFB7F736": {
		"Type": "AWS::EC2::SecurityGroupIngress",
		"Properties": {
		  "IpProtocol": "tcp",
		  "Description": "Load balancer to target",
		  "FromPort": 3030,
		  "GroupId": {
			"Fn::GetAtt": [
			  "ASGInstanceSecurityGroup0525485D",
			  "GroupId"
			]
		  },
		  "SourceSecurityGroupId": {
			"Fn::GetAtt": [
			  "LBSecurityGroup8A41EA2B",
			  "GroupId"
			]
		  },
		  "ToPort": 3030
		},
		"Metadata": {
		  "aws:cdk:path": "AlbEc2Stack/ASG/InstanceSecurityGroup/from AlbEc2StackLBSecurityGroup7075EADF:3030"
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
		  "aws:cdk:path": "AlbEc2Stack/ASG/InstanceProfile"
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
				  "Ref": "InstanceClass"
				},
				".",
				{
				  "Ref": "InstanceSize"
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
				  "#!/bin/bash\nnest-secrets --prefix /",
				  {
					"Ref": "App"
				  },
				  "/",
				  {
					"Fn::FindInMap": [
					  "stages",
					  {
						"Ref": "Stage"
					  },
					  "lower"
					]
				  },
				  " > .env\naws s3 cp s3://",
				  {
					"Ref": "S3Bucket"
				  },
				  "/",
				  {
					"Ref": "S3Key"
				  },
				  " app.tar.gz\ndocker load < app.tar.gz\ndocker run             --env-file .env             -p 3030:3030             --log-driver=awslogs             --log-opt awslogs-group=",
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
				  "             --log-opt awslogs-create-group=true             ",
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
		  "aws:cdk:path": "AlbEc2Stack/ASG/LaunchConfig"
		}
	  },
	  "ASG46ED3070": {
		"Type": "AWS::AutoScaling::AutoScalingGroup",
		"Properties": {
		  "MaxSize": {
			"Ref": "MaxCapacity"
		  },
		  "MinSize": {
			"Ref": "MinCapacity"
		  },
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
			  "Value": "AlbEc2Stack/ASG"
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
		  "TargetGroupARNs": [
			{
			  "Ref": "LBListenerTargetGroupF04FCF6D"
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
		  "aws:cdk:path": "AlbEc2Stack/ASG/ASG"
		}
	  },
	  "ASGScalingPolicyGTCPUF089F755": {
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
			"TargetValue": {
			  "Ref": "TargetCPU"
			}
		  }
		},
		"Metadata": {
		  "aws:cdk:path": "AlbEc2Stack/ASG/ScalingPolicyGTCPU/Resource"
		}
	  },
	  "LB8A12904C": {
		"Type": "AWS::ElasticLoadBalancingV2::LoadBalancer",
		"Properties": {
		  "Name": {
			"Fn::Join": [
			  "",
			  [
				{
				  "Ref": "App"
				},
				"-",
				{
				  "Ref": "Stage"
				}
			  ]
			]
		  },
		  "Scheme": "internet-facing",
		  "SecurityGroups": [
			{
			  "Fn::GetAtt": [
				"LBSecurityGroup8A41EA2B",
				"GroupId"
			  ]
			}
		  ],
		  "Subnets": {
			"Ref": "Subnets"
		  },
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
		  ],
		  "Type": "application"
		},
		"Metadata": {
		  "aws:cdk:path": "AlbEc2Stack/LB/Resource"
		}
	  },
	  "LBSecurityGroup8A41EA2B": {
		"Type": "AWS::EC2::SecurityGroup",
		"Properties": {
		  "GroupDescription": "Automatically created Security Group for ELB AlbEc2StackLB93E8F97D",
		  "SecurityGroupIngress": [
			{
			  "CidrIp": "0.0.0.0/0",
			  "Description": "Allow from anyone on port 443",
			  "FromPort": 443,
			  "IpProtocol": "tcp",
			  "ToPort": 443
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
		  ],
		  "VpcId": {
			"Ref": "VpcId"
		  }
		},
		"Metadata": {
		  "aws:cdk:path": "AlbEc2Stack/LB/SecurityGroup/Resource"
		}
	  },
	  "LBSecurityGrouptoAlbEc2StackASGInstanceSecurityGroupEE06B44E303026BD048F": {
		"Type": "AWS::EC2::SecurityGroupEgress",
		"Properties": {
		  "GroupId": {
			"Fn::GetAtt": [
			  "LBSecurityGroup8A41EA2B",
			  "GroupId"
			]
		  },
		  "IpProtocol": "tcp",
		  "Description": "Load balancer to target",
		  "DestinationSecurityGroupId": {
			"Fn::GetAtt": [
			  "ASGInstanceSecurityGroup0525485D",
			  "GroupId"
			]
		  },
		  "FromPort": 3030,
		  "ToPort": 3030
		},
		"Metadata": {
		  "aws:cdk:path": "AlbEc2Stack/LB/SecurityGroup/to AlbEc2StackASGInstanceSecurityGroupEE06B44E:3030"
		}
	  },
	  "LBListener49E825B4": {
		"Type": "AWS::ElasticLoadBalancingV2::Listener",
		"Properties": {
		  "DefaultActions": [
			{
			  "TargetGroupArn": {
				"Ref": "LBListenerTargetGroupF04FCF6D"
			  },
			  "Type": "forward"
			}
		  ],
		  "LoadBalancerArn": {
			"Ref": "LB8A12904C"
		  },
		  "Port": 443,
		  "Protocol": "HTTPS",
		  "Certificates": [
			{
			  "CertificateArn": {
				"Ref": "CertificateArn"
			  }
			}
		  ]
		},
		"Metadata": {
		  "aws:cdk:path": "AlbEc2Stack/LB/Listener/Resource"
		}
	  },
	  "LBListenerTargetGroupF04FCF6D": {
		"Type": "AWS::ElasticLoadBalancingV2::TargetGroup",
		"Properties": {
		  "HealthCheckIntervalSeconds": 30,
		  "HealthCheckPath": "/healthcheck",
		  "HealthCheckTimeoutSeconds": 10,
		  "HealthyThresholdCount": 2,
		  "Port": 3030,
		  "Protocol": "HTTP",
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
		  ],
		  "TargetType": "instance",
		  "UnhealthyThresholdCount": 5,
		  "VpcId": {
			"Ref": "VpcId"
		  }
		},
		"Metadata": {
		  "aws:cdk:path": "AlbEc2Stack/LB/Listener/TargetGroup/Resource"
		}
	  }
	}
  }
`
