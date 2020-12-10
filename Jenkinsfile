pipeline {
  agent { label 'ubuntu_docker_label' }
  tools {
    go "Go 1.15"
  }
  environment {
    PROJECT     = "src/github.com/infobloxopen/schema-registry-helper"
    GOPATH      = "$WORKSPACE"
  }
  options {
    checkoutToSubdirectory('src/github.com/infobloxopen/schema-registry-helper')
  }
  stages {
    stage('Test') {
      steps {
        sh 'cd $PROJECT && make test'
      }
    }
    stage('Build image') {
      steps {
        sh 'cd $PROJECT && make examples'
        sh 'cd $PROJECT && git diff --exit-code'
      }
    }
  }
}
