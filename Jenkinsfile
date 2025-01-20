pipeline {
    agent any

    environment {
        GO_VERSION = 'latest' // Set to the latest Go version
        DOCKER_IMAGE = 'golang:latest' // Use the latest Go Docker image
        DOCKER_HOST = 'tcp://docker:2376' // Docker host, adjust if necessary
        DOCKER_COMPOSE = '/usr/libexec/docker/cli-plugins/docker-compose' // Docker Compose path
    }

    tools {
        go 'Go 1.19' // Set Go version installed in Jenkins
    }

    stages {
        stage('Checkout') {
            steps {
                // Checkout the code from the repository
                checkout scm
            }
        }

        stage('Install Dependencies') {
            steps {
                // Install dependencies (e.g., using Go modules)
                sh '''
                    go mod tidy
                '''
            }
        }

        stage('Build Application') {
            steps {
                // Build the Go application
                sh '''
                    go build -o myapp .
                '''
            }
        }

        stage('Docker Build') {
            steps {
                // Build the Docker image for the Go app
                script {
                    docker.build("myapp:${BUILD_NUMBER}")
                }
            }
        }

        stage('Docker Push') {
            steps {
                withCredentials([usernamePassword(credentialsId: 'docker-hub-credentials', usernameVariable: 'DOCKER_USERNAME', passwordVariable: 'DOCKER_PASSWORD')]) {
                    // Login to Docker Hub and push the image
                    sh '''
                        echo $DOCKER_PASSWORD | docker login -u $DOCKER_USERNAME --password-stdin
                        docker push myapp:${BUILD_NUMBER}
                    '''
                }
            }
        }

        stage('Deploy') {
            steps {
                // Deploy the application using Docker Compose
                sh '''
                    ${DOCKER_COMPOSE} down || true
                    ${DOCKER_COMPOSE} up -d
                '''
            }
        }
    }
}
