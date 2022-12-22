@Library('jenkins-library@opensource-release') _
dockerImagePipeline(
  script: this,
  service: 'kubeslice-router-sidecar',
  dockerfile: 'Dockerfile',
  buildContext: '.',
  buildArguments: [PLATFORM:"amd64"]
  
)
