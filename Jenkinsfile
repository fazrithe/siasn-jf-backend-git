pipeline {
    parameters {
        string(name: 'PRODUCTION_NAMESPACE',       description: 'Production Namespace',                 defaultValue: 'siasn2021')

        string(name: 'DEVELOPMENT_NAMESPACE',      description: 'Development Namespace',                defaultValue: 'siasn2021-training')

        string(name: 'DOCKER_IMAGE_NAME',          description: 'Docker Image Name',                    defaultValue: 'siasn2021-jf-backend')
    }
	
    agent any
       triggers {
           pollSCM(env.BRANCH_NAME == 'master' ? '* * * * *' : '* * * * *')
    }

    
    
    stages {

        stage('Checkout SCM') {
            steps {
                
                script{
                      
                      while(currentBuild.rawBuild.getPreviousBuildInProgress() != null) {
                       currentBuild.rawBuild.getPreviousBuildInProgress().doKill()
                     }

                    sh 'rm -Rf *'

                }
                checkout scm
                script {
                    echo "get COMMIT_ID"
                    sh 'echo -n $(git rev-parse --short HEAD) > ./commit-id'
                    commitId = readFile('./commit-id')
                }
                stash(name: 'ws', includes:'**,./commit-id') // stash this current 
            }
        }

        stage('Initialize') {
            steps {
                script{
                            if ( env.BRANCH_NAME == 'master' ){
				    projectKubernetes= "${params.PRODUCTION_NAMESPACE}"
                                envStage = "production"
                            }else if ( env.BRANCH_NAME == 'development'){
				projectKubernetes= "${params.DEVELOPMENT_NAMESPACE}"
                                envStage = "development"
                    }
                    echo "${projectKubernetes}"
                    
                } 
                
            }
        }


        stage('SonarQube') {
            steps {
                
                script{
                    
                    
                    def scannerHome = tool 'sonarqube4.5' ;
                            withSonarQubeEnv('SonarQube') {
                                sh "${scannerHome}/bin/sonar-scanner"
                      }
                    
                }
                
                
            }
        }

        stage('Build Docker') {
            steps {
                
                script{
                    
                    sh "docker build --rm --no-cache --pull -t ${params.DOCKER_IMAGE_NAME}:${BUILD_NUMBER}-${commitId} ."
                    
                }
            }
        }

       stage('Deploy') {
            steps {
		    script{
                	   echo "Login Docker Registry"
    			withCredentials([usernamePassword(credentialsId: 'registry', usernameVariable: 'username', passwordVariable: 'password')]) {

        			sh "docker login harbor.bkn.go.id -u admin -p ${password}"

                 imagefinal = "harbor.bkn.go.id/${projectKubernetes}/${params.DOCKER_IMAGE_NAME}"
				 sh "docker tag ${params.DOCKER_IMAGE_NAME}:${BUILD_NUMBER}-${commitId} ${imagefinal}:latest"
				 sh  "docker push  ${imagefinal}:latest"
                 sh "docker tag ${params.DOCKER_IMAGE_NAME}:${BUILD_NUMBER}-${commitId} ${imagefinal}:prod-code-${BUILD_NUMBER}"
				 sh  "docker push ${imagefinal}:prod-code-${BUILD_NUMBER}"
				 //deploy project

                 if (env.BRANCH_NAME == 'master'){
				 //deploy project ke prod
					sh "KUBECONFIG=/home/bkn/.kube/workground-okd4  oc -n  siasn2021 set image deploymentconfig/${params.DOCKER_IMAGE_NAME} ${params.DOCKER_IMAGE_NAME}=${imagefinal}:prod-code-${BUILD_NUMBER}"
				 }else if (env.BRANCH_NAME == 'development'){
				 //deploy ke training
					sh "KUBECONFIG=/home/bkn/.kube/workground-okd4  oc -n  siasn2021-training set image deploymentconfig/${params.DOCKER_IMAGE_NAME} ${params.DOCKER_IMAGE_NAME}=${imagefinal}:prod-code-${BUILD_NUMBER}"
				 }
                 sh "docker rmi ${params.DOCKER_IMAGE_NAME}:${BUILD_NUMBER}-${commitId}"
                 sh "docker rmi ${imagefinal}:prod-code-${BUILD_NUMBER}"

    			}
		    }
		    
            }
        }


    }


post {
        always{
          
                  script{
                                if ( currentBuild.currentResult == 'SUCCESS' ) {
                                        textMessage = "\u2600 Jenkins  ${JOB_NAME}-${BUILD_NUMBER} SUCCESS"
                        
                                        withCredentials([string(credentialsId: 'token_telegram', variable: 'TELEGRAM_TOKEN')]) {
                                                                sh "curl -s -X POST 'https://api.telegram.org/${TELEGRAM_TOKEN}/sendMessage?chat_id=-340782909&text=${textMessage}&parse_mode=HTML'"
                                        }
                        
                               }
                          else if( currentBuild.currentResult == 'FAILURE' ) {
                            textMessage = "\u26c8 Jenkins ${JOB_NAME}-${BUILD_NUMBER} FAILED"  
            
                            withCredentials([string(credentialsId: 'token_telegram', variable: 'TELEGRAM_TOKEN')]) {
                                                    sh "curl -s -X POST 'https://api.telegram.org/${TELEGRAM_TOKEN}/sendMessage?chat_id=-340782909&text=${textMessage}&parse_mode=HTML'"
                              }
                    
                          }
                           sh  "docker rmi -f  ${imagefinal}:latest"
                       }
          	 sh  "docker rmi -f  ${imagefinal}:latest"
          
        }
	
       }



}