import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecr from 'aws-cdk-lib/aws-ecr';
import * as eks from 'aws-cdk-lib/aws-eks';
import { KubectlV31Layer } from '@aws-cdk/lambda-layer-kubectl-v31';
import { Construct } from 'constructs';

export interface KblPlatformStackProps extends cdk.StackProps {
  /** EKS Kubernetes version */
  readonly kubernetesVersion?: eks.KubernetesVersion;
  /** Initial managed node group instance type */
  readonly nodeInstanceType?: ec2.InstanceType;
}

/**
 * AWS foundation for KBL Compute Engine.
 *
 * Phase 22 scaffold: ECR image repos + VPC + EKS cluster with a small node group.
 * MSK (Kafka multiverse), FSx/Lustre data staging, and Helm chart install are deferred.
 *
 * Deploy lab manifests from ../../lab/kustomize after pushing images to ECR.
 */
export class KblPlatformStack extends cdk.Stack {
  public readonly cluster: eks.Cluster;
  public readonly repositories: Record<string, ecr.Repository>;

  constructor(scope: Construct, id: string, props: KblPlatformStackProps = {}) {
    super(scope, id, props);

    const k8sVersion = props.kubernetesVersion ?? eks.KubernetesVersion.V1_31;
    const nodeType = props.nodeInstanceType ?? ec2.InstanceType.of(
      ec2.InstanceClass.T3,
      ec2.InstanceSize.MEDIUM,
    );

    const vpc = new ec2.Vpc(this, 'KblVpc', {
      maxAzs: 2,
      natGateways: 1,
      subnetConfiguration: [
        { name: 'public', subnetType: ec2.SubnetType.PUBLIC, cidrMask: 24 },
        { name: 'private', subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS, cidrMask: 24 },
      ],
    });

    const imageNames = [
      'kbl-controller',
      'kbl-tsdb',
      'kbl-domino-runner',
      'kbl-domino-runner-julia',
    ] as const;

    this.repositories = {} as Record<string, ecr.Repository>;
    for (const name of imageNames) {
      this.repositories[name] = new ecr.Repository(this, `${name}Repo`, {
        repositoryName: name,
        imageScanOnPush: true,
        removalPolicy: cdk.RemovalPolicy.RETAIN,
      });
    }

    this.cluster = new eks.Cluster(this, 'KblEks', {
      vpc,
      version: k8sVersion,
      kubectlLayer: new KubectlV31Layer(this, 'KubectlLayer'),
      defaultCapacity: 0,
      clusterLogging: [
        eks.ClusterLoggingTypes.API,
        eks.ClusterLoggingTypes.AUDIT,
        eks.ClusterLoggingTypes.AUTHENTICATOR,
      ],
    });

    this.cluster.addNodegroupCapacity('DefaultNg', {
      instanceTypes: [nodeType],
      minSize: 1,
      maxSize: 3,
      desiredSize: 2,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });

    new cdk.CfnOutput(this, 'ClusterName', { value: this.cluster.clusterName });
    new cdk.CfnOutput(this, 'ClusterEndpoint', { value: this.cluster.clusterEndpoint });
    for (const name of imageNames) {
      new cdk.CfnOutput(this, `${name}EcrUri`, {
        value: this.repositories[name].repositoryUri,
      });
    }

    cdk.Tags.of(this).add('project', 'kbl-compute');
  }
}
