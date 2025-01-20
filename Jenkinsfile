pipeline {
    agent any

    environment {
        DOCKER_IMAGE = 'golang:latest' // Use the latest Go Docker image
        DOCKER_HOST = 'tcp://docker:2376' // Docker host, adjust if necessary
        DOCKER_COMPOSE = '/usr/libexec/docker/cli-plugins/docker-compose' // Docker Compose path
        IMAGE_NAME = 'myapp' // Name of the Docker image
    }

    stages {
        stage('Checkout') {
            steps {
                // Checkout the code from the repository
                checkout scm
            }
        }

        stage('Docker Build') {
            steps {
                // Build the Docker image from the Dockerfile in the repository
                script {
                    docker.build("${IMAGE_NAME}:${BUILD_NUMBER}")
                }
            }
        }

        stage('Docker Push') {
            steps {
                withCredentials([usernamePassword(credentialsId: 'docker-hub-credentials', usernameVariable: 'DOCKER_USERNAME', passwordVariable: 'DOCKER_PASSWORD')]) {
                    // Login to Docker Hub and push the image
                    sh '''
                        echo $DOCKER_PASSWORD | docker login -u $DOCKER_USERNAME --password-stdin
                        docker push ${IMAGE_NAME}:${BUILD_NUMBER}
                    '''
                }
            }
        }
    }
}
