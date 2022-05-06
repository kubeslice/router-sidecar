@Library('jenkins-library@opensource') _
dockerImagePipeline(
  script: this,
  service: 'aveshasystems/kubeslice-router-sidecar',
  dockerfile: 'Dockerfile',
  buildContext: '.',
  buildArguments: [PLATFORM:"amd64"]
)
