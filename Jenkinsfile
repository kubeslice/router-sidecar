@Library('jenkins-library@opensource-nexus-poc') _
dockerImagePipeline(
  script: this,
  service: 'kubeslice-router-sidecar',
  dockerfile: 'Dockerfile',
  buildContext: '.',
  buildArguments: [PLATFORM:"amd64"]
)
