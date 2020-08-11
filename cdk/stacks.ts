import { AlbEc2Stack } from "./stacks/AlbEc2Stack";
import * as cdk from "@aws-cdk/core";

const props = { env: { region: "eu-west-1" } };

const app = new cdk.App();
// tslint:disable-next-line: no-unused-expression
new AlbEc2Stack(app, "AlbEc2Stack", props);
