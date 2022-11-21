// Copyright (c) HashiCorp, Inc
// SPDX-License-Identifier: MPL-2.0
import { Construct } from "constructs";
import { App, TerraformStack, CloudBackend, NamedCloudWorkspace } from "cdktf";
import * as google from '@cdktf/provider-google';

const project = 'studious-journey';
const region = 'asia-northeast1';
const repository = 'studious-journey';

class MyStack extends TerraformStack {
  constructor(scope: Construct, id: string) {
    super(scope, id);

    new google.provider.GoogleProvider(this, 'google', {
      project,
      region,
    });

    new google.artifactRegistryRepository.ArtifactRegistryRepository(this, 'artifact_registry', {
      format: 'docker',
      location: region,
      repositoryId: repository,
    });

    new google.cloudbuildTrigger.CloudbuildTrigger(this, 'build_trigger', {
      filename: 'cloudbuild.yaml',
      github: {
        owner: 'hsmtkk',
        name: repository,
        push: {
          branch: 'main',
        },
      },
    });

    const cloud_run_service_account = new google.serviceAccount.ServiceAccount(this, 'cloud_run_service_account', {
      accountId: 'cloud-runner',
      displayName: 'service account for Cloud Run',
    });

    new google.projectIamBinding.ProjectIamBinding(this, 'cloud_run_to_cloud_monitoring', {
      members: [`serviceAccount:${cloud_run_service_account.email}`],
      project,
      role: 'roles/monitoring.metricWriter',
    });

    const cloud_run_no_auth_policy = new google.dataGoogleIamPolicy.DataGoogleIamPolicy(this, 'cloud_run_no_auth_policy', {
      binding: [{
        role: 'roles/run.invoker',
        members: ['allUsers'],
      }],
    });

    const test_service = new google.cloudRunService.CloudRunService(this, 'test_service', {
      location: region,
      name: 'test-service',
      template: {
        spec: {
          containers: [{
            image: 'us-docker.pkg.dev/cloudrun/container/hello',
          }],
          serviceAccountName: cloud_run_service_account.email,
        },
      },
    });

    new google.cloudRunServiceIamPolicy.CloudRunServiceIamPolicy(this, 'test_service_no_auth', {
      location: region,
      policyData: cloud_run_no_auth_policy.policyData,
      service: test_service.name,
    });
  }
}

const app = new App();
const stack = new MyStack(app, "studious-journey");
new CloudBackend(stack, {
  hostname: "app.terraform.io",
  organization: "hsmtkkdefault",
  workspaces: new NamedCloudWorkspace("studious-journey")
});
app.synth();
