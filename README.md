# Nest

A basic tool to help build and deploy Guardian services.

## Motivation

Creating a new service should be simple. Similarly, updating a service or indeed
multiple services should be simple.

Nest aims to do this by providing a few supported deployment patterns. If you
stick to these, you can benefit from Nest updates over time - for example,
logging improvements, network security fixes, and so on.

To achieve this, Docker is used as the unit of deployment.

## Usage

    $ nest help // to see usage info

At the moment, only one deployment type is supported: `ec2-service`. This is a
simple recipe to deploy instances to an ASG.

To use, create a `nest.json` file in the root of your repository:

```
{
    "app": "contributions-service",
    "stack": "frontend",
    "vcsURL": "https://github.com/guardian/contributions-service",
    "deploymentType": "ec2-service",
    "artifactBucket": "aws-frontend-artifacts"
}
```

The usual Riffraff rules apply - your Riffraff user will need to have permission
to write to your artifact bucket for example.

Make sure you have a `Dockerfile` in the root of your repository too, which
builds your app. It should respect a PORT environment variable and listen on
that.

Logging is provided by Cloudwatch Logs (which can be forwarded to Kinesis/ELK if
you like).

Extra reading:

https://riffraff.gutools.co.uk/docs/reference/s3-artifact-layout.md.
