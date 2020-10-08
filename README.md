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

At the moment, only one deployment type is supported: `alb-ec2-service`. This is
a simple recipe to deploy instances to an ASG. Out of the box it provides:

- a public load balancer, pointing to instances living in an autoscaling group
- [ssm](https://github.com/guardian/ssm-scala) compatibility (via the tunnel
  approach)
- logging (Cloudwatch Logs)
- secrets/config (via Parameter Store)

See the specific sections below for details about these.

To use this recipe, simply make sure you have a `Dockerfile` in the root of your
repository, which builds your app. It should start up your service on the port
of the `PORT` environment variable (which Nest will specify).

Then, grab the latest release of Nest suitable for your platform ('darwin' if
Mac) and run:

    $ nest init
    $ nest build
    $ nest upload

Nest's `init` command will create a `nest.json` config file, which it is
recommended to check in as it is required by the main nest commands.

For additional help, see:

    $ nest help

The usual Riffraff rules apply - your Riffraff user will need to have permission
to write to your artifact bucket for example.

## Custom resources

Nest provides a useful mechanism to provision core resources like EC2 instances
and ALBs, but there are likely to be other, non-standard resources you want too.
These might be: custom IAM policies, alarms, or a database like Elasticsearch.
To manage these, the `alb-ec2-service` recipe supports an optional
`customCloudformation` field; set it to the relative path of a valid
Cloudformation template, and this will be deployed as part of your build as a
separate Cloudformation stack. The naming convention for it is
`[STACK]-[app]-custom-[STAGE]` or `[STACK]-[custom-name]-custom-[STAGE]` if you
have used the `cloudformationStackName` Nest config parameter to customise
things.

## Logging

Logging is provided by Cloudwatch Logs (which can be forwarded to Kinesis/ELK if
you like). Simply log to standout out/error. We recommend a JSON format.

## Secrets/config

Nest uses [nest-secrets](https://github.com/guardian/nest-secrets) to pass
configuration into your container as environment variables. Simply prefix your
config in Parameter Store with `/${app}/{stage_lower}` and they will be provided
on startup to your app as environment variables.

Note, `/` and `.` in parameter names are converted to `_`. See the
`nest-secrets` README for more info here.

## Local development

TODO.

## Extra reading

If you want to understand the Riffraff deployment model better, we recommend
reading:

https://riffraff.gutools.co.uk/docs/reference/s3-artifact-layout.md.
