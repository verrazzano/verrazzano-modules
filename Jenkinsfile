// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

def DOCKER_IMAGE_TAG
def SKIP_ACCEPTANCE_TESTS = false
def SKIP_TRIGGERED_TESTS = false
def SUSPECT_LIST = ""
def VERRAZZANO_DEV_VERSION = ""
def tarfilePrefix=""
def storeLocation=""

def agentLabel = env.JOB_NAME.contains('main') ? "phx-large" : "large"

pipeline {
    options {
        skipDefaultCheckout true
        copyArtifactPermission('*');
        timestamps ()
    }

    agent {
       docker {
            image "${RUNNER_DOCKER_IMAGE}"
            args "${RUNNER_DOCKER_ARGS}"
            registryUrl "${RUNNER_DOCKER_REGISTRY_URL}"
            registryCredentialsId 'ocir-pull-and-push-account'
            label "${agentLabel}"
        }
    }

    parameters {
        booleanParam (description: 'Whether to perform a scan of the built images', name: 'PERFORM_SCAN', defaultValue: false)
    }

    environment {
        TEST_ENV = "JENKINS"
        CLEAN_BRANCH_NAME = "${env.BRANCH_NAME.replace("/", "%2F")}"

        DOCKER_MODULE_CI_IMAGE_NAME = 'verrazzano-module-operator-jenkins'
        DOCKER_MODULE_PUBLISH_IMAGE_NAME = 'verrazzano-module-operator'
        DOCKER_MODULE_IMAGE_NAME = "${env.BRANCH_NAME ==~ /^release-.*/ || env.BRANCH_NAME == 'main' ? env.DOCKER_MODULE_PUBLISH_IMAGE_NAME : env.DOCKER_MODULE_CI_IMAGE_NAME}"

        DOCKER_HELM_CI_IMAGE_NAME = 'verrazzano-helm-operator-jenkins'
        DOCKER_HELM_PUBLISH_IMAGE_NAME = 'verrazzano-helm-operator'
        DOCKER_HELM_IMAGE_NAME = "${env.BRANCH_NAME ==~ /^release-.*/ || env.BRANCH_NAME == 'main' ? env.DOCKER_HELM_PUBLISH_IMAGE_NAME : env.DOCKER_HELM_CI_IMAGE_NAME}"

        DOCKER_CALICO_CI_IMAGE_NAME = 'verrazzano-calico-operator-jenkins'
        DOCKER_CALICO_PUBLISH_IMAGE_NAME = 'verrazzano-calico-operator'
        DOCKER_CALICO_IMAGE_NAME = "${env.BRANCH_NAME ==~ /^release-.*/ || env.BRANCH_NAME == 'main' ? env.DOCKER_CALICO_PUBLISH_IMAGE_NAME : env.DOCKER_CALICO_CI_IMAGE_NAME}"

        CREATE_LATEST_TAG = "${env.BRANCH_NAME == 'main' ? '1' : '0'}"
        GOPATH = '/home/opc/go'
        GO_REPO_PATH = "${GOPATH}/src/github.com/verrazzano"
        GIT_REPO_DIR = "verrazzano-modules"
        DOCKER_CREDS = credentials('github-packages-credentials-rw')
        DOCKER_EMAIL = credentials('github-packages-email')
        DOCKER_REPO = 'ghcr.io'
        DOCKER_NAMESPACE = 'verrazzano'
        NETRC_FILE = credentials('netrc')
        GITHUB_PKGS_CREDS = credentials('github-packages-credentials-rw')
        SERVICE_KEY = credentials('PAGERDUTY_SERVICE_KEY')

        POST_DUMP_FAILED_FILE = "${WORKSPACE}/post_dump_failed_file.tmp"
        TESTS_EXECUTED_FILE = "${WORKSPACE}/tests_executed_file.tmp"
        //KUBECONFIG = "${WORKSPACE}/test_kubeconfig"
        OCR_CREDS = credentials('ocr-pull-and-push-account')
        OCR_REPO = 'container-registry.oracle.com'
        IMAGE_PULL_SECRET = 'verrazzano-container-registry'

        // used for console artifact capture on failure
        JENKINS_READ = credentials('jenkins-auditor')

        OCI_CLI_AUTH="instance_principal"
        OCI_OS_NAMESPACE = credentials('oci-os-namespace')
        OCI_OS_ARTIFACT_BUCKET="build-failure-artifacts"
        OCI_OS_BUCKET="verrazzano-builds"
        OCI_OS_COMMIT_BUCKET="verrazzano-builds-by-commit"
        OCI_OS_REGION="us-phoenix-1"

        // used to emit metrics
        PROMETHEUS_CREDENTIALS = credentials('prometheus-credentials')

        // used to write to object storage, or fail build if UT coverage does not pass
        //FAIL_BUILD_COVERAGE = "${params.FAIL_IF_COVERAGE_DECREASED}"
        //UPLOAD_UT_COVERAGE = "${params.UPLOAD_UNIT_TEST_COVERAGE}"
    }

    stages {
        stage('Clean workspace and checkout') {
            steps {
                sh """
                    echo "${NODE_LABELS}"
                """

                script {
                    def scmInfo = checkout scm
                    env.GIT_COMMIT = scmInfo.GIT_COMMIT
                    env.GIT_BRANCH = scmInfo.GIT_BRANCH
                    echo "SCM checkout of ${env.GIT_BRANCH} at ${env.GIT_COMMIT}"
                }
                sh """
                    cp -f "${NETRC_FILE}" $HOME/.netrc
                    chmod 600 $HOME/.netrc
                """

                script {
                    try {
                        sh """
                            echo "${DOCKER_CREDS_PSW}" | docker login ${env.DOCKER_REPO} -u ${DOCKER_CREDS_USR} --password-stdin
                        """
                    } catch(error) {
                        echo "docker login failed, retrying after sleep"
                        retry(4) {
                            sleep(30)
                            sh """
                            echo "${DOCKER_CREDS_PSW}" | docker login ${env.DOCKER_REPO} -u ${DOCKER_CREDS_USR} --password-stdin
                            """
                        }
                    }
                }
                moveContentToGoRepoPath()

                script {
                    def props = readProperties file: '.verrazzano-development-version'
                    VERRAZZANO_DEV_VERSION = props['verrazzano-development-version']
                    TIMESTAMP = sh(returnStdout: true, script: "date +%Y%m%d%H%M%S").trim()
                    SHORT_COMMIT_HASH = sh(returnStdout: true, script: "echo $env.GIT_COMMIT | head -c 8")
                    env.VERRAZZANO_VERSION = "${VERRAZZANO_DEV_VERSION}"
                    if (!"${env.GIT_BRANCH}".startsWith("release-")) {
                        env.VERRAZZANO_VERSION = "${env.VERRAZZANO_VERSION}-${env.BUILD_NUMBER}+${SHORT_COMMIT_HASH}"
                    }
                    DOCKER_IMAGE_TAG = "v${VERRAZZANO_DEV_VERSION}-${TIMESTAMP}-${SHORT_COMMIT_HASH}"
                    // update the description with some meaningful info
                    currentBuild.description = SHORT_COMMIT_HASH + " : " + env.GIT_COMMIT
                    def currentCommitHash = env.GIT_COMMIT
                    def commitList = getCommitList()
                    withCredentials([file(credentialsId: 'jenkins-to-slack-users', variable: 'JENKINS_TO_SLACK_JSON')]) {
                        def userMappings = readJSON file: JENKINS_TO_SLACK_JSON
                        SUSPECT_LIST = getSuspectList(commitList, userMappings)
                        echo "Suspect list: ${SUSPECT_LIST}"
                    }
                }
            }
        }

        stage('Check Repo Clean') {
            steps {
                checkRepoClean()
            }
        }

        stage('Parallel Build, Test, and Compliance') {
            parallel {
                stage('Build Images and Save Generated Files') {
                    when { not { buildingTag() } }
                    steps {
                        script {
                            buildImages("${DOCKER_IMAGE_TAG}")
                            generateOperatorYaml("${DOCKER_IMAGE_TAG}")
                        }
                    }
                    post {
                        success {
                            echo "Saving generated files"
                            saveGeneratedFiles()
                            script {
                                archiveArtifacts artifacts: "generated/*.yaml,generated/*.tgz,verrazzano_images.txt", allowEmptyArchive: true
                            }
                        }
                    }
                }

                stage('Quality, Compliance Checks, and Unit Tests') {
                   when { not { buildingTag() } }
                   steps {
                       sh """
                           echo "Not implemented"
                           #cd ${GO_REPO_PATH}/${GIT_REPO_DIR}
                           #make precommit
                           #make unit-test-coverage-ratcheting
                       """
                   }
                   post {
                       always {
                           sh """
                               cd ${GO_REPO_PATH}/${GIT_REPO_DIR}
                               #cp coverage.html ${WORKSPACE}
                               #cp coverage.xml ${WORKSPACE}
                               #build/copy-junit-output.sh ${WORKSPACE}
                           """
                           archiveArtifacts artifacts: '**/coverage.html', allowEmptyArchive: true
                           junit testResults: '**/*test-result.xml', allowEmptyResults: true
                           cobertura(coberturaReportFile: 'coverage.xml',
                                   enableNewApi: true,
                                   autoUpdateHealth: false,
                                   autoUpdateStability: false,
                                   failUnstable: true,
                                   failUnhealthy: true,
                                   failNoReports: true,
                                   onlyStable: false,
                                   fileCoverageTargets: '100, 0, 0',
                                   lineCoverageTargets: '68, 68, 68',
                                   packageCoverageTargets: '100, 0, 0',
                           )
                       }
                   }
                }
            }

        }

       //stage('Scan Image') {
       //     when {
       //        allOf {
       //            not { buildingTag() }
       //            expression {params.PERFORM_SCAN == true}
       //        }
       //     }
       //     steps {
       //         script {
       //             scanContainerImage "${env.DOCKER_REPO}/${env.DOCKER_NAMESPACE}/${DOCKER_PLATFORM_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
       //         }
       //     }
       //     post {
       //         always {
       //             archiveArtifacts artifacts: '**/scanning-report*.json', allowEmptyArchive: true
       //         }
       //     }
       //}
    }

    post {
        always {
            archiveArtifacts artifacts: '**/coverage.html,**/logs/**,**/verrazzano_images.txt,**/*cluster-snapshot*/**', allowEmptyArchive: true
            junit testResults: '**/*test-result.xml', allowEmptyResults: true
        }
        failure {
            sh """
                curl -k -u ${JENKINS_READ_USR}:${JENKINS_READ_PSW} -o ${WORKSPACE}/build-console-output.log ${BUILD_URL}consoleText
            """
            archiveArtifacts artifacts: '**/build-console-output.log', allowEmptyArchive: true
            sh """
                curl -k -u ${JENKINS_READ_USR}:${JENKINS_READ_PSW} -o archive.zip ${BUILD_URL}artifact/*zip*/archive.zip
                oci --region us-phoenix-1 os object put --force --namespace ${OCI_OS_NAMESPACE} -bn ${OCI_OS_ARTIFACT_BUCKET} --name ${env.JOB_NAME}/${env.BRANCH_NAME}/${env.BUILD_NUMBER}/archive.zip --file archive.zip
                rm archive.zip
            """
            script {
                if (isPagerDutyEnabled() && (env.JOB_NAME == "verrazzano-modules/main" || env.JOB_NAME ==~ "verrazzano-modules/release-1.*")) {
                    pagerduty(resolve: false, serviceKey: "$SERVICE_KEY", incDescription: "Verrazzano: ${env.JOB_NAME} - Failed", incDetails: "Job Failed - \"${env.JOB_NAME}\" build: ${env.BUILD_NUMBER}\n\nView the log at:\n ${env.BUILD_URL}\n\nBlue Ocean:\n${env.RUN_DISPLAY_URL}")
                }
                if (env.JOB_NAME == "verrazzano-modules/main" || env.JOB_NAME ==~ "verrazzano-modules/release-1.*" || env.BRANCH_NAME ==~ "mark/*") {
                    slackSend ( channel: "$SLACK_ALERT_CHANNEL", message: "Job Failed - \"${env.JOB_NAME}\" build: ${env.BUILD_NUMBER}\n\nView the log at:\n ${env.BUILD_URL}\n\nBlue Ocean:\n${env.RUN_DISPLAY_URL}\n\nSuspects:\n${SUSPECT_LIST}" )
                }
            }
        }
        cleanup {
            deleteDir()
        }
    }
}

def isPagerDutyEnabled() {
    // this controls whether PD alerts are enabled
    if (NOTIFY_PAGERDUTY_MAINJOB_FAILURES.equals("true")) {
        echo "Pager-Duty notifications enabled via global override setting"
        return true
    }
    return false
}

// Called in Stage Clean workspace and checkout steps
def moveContentToGoRepoPath() {
    sh """
        rm -rf ${GO_REPO_PATH}/${GIT_REPO_DIR}
        mkdir -p ${GO_REPO_PATH}/${GIT_REPO_DIR}
        tar cf - . | (cd ${GO_REPO_PATH}/${GIT_REPO_DIR}/ ; tar xf -)
    """
}

def checkRepoClean() {
    sh """
        echo "Not implemented"
        #cd ${GO_REPO_PATH}/${GIT_REPO_DIR}
        #echo 'Check for forgotten manifest/generate actions...'
        #(cd platform-operator; make check-repo-clean)
        #(cd application-operator; make check-repo-clean)
        #(cd cluster-operator; make check-repo-clean)
    """
}

// Called in Stage Build steps
// Makes target docker push for application/platform operator and analysis
def buildImages(dockerImageTag) {
    sh """
        cd ${GO_REPO_PATH}/${GIT_REPO_DIR}
        echo 'Building container images...'
        make docker-push \
            DOCKER_REPO=${env.DOCKER_REPO} DOCKER_NAMESPACE=${env.DOCKER_NAMESPACE} \
            DOCKER_IMAGE_TAG=${dockerImageTag} \
            VERRAZZANO_MODULE_OPERATOR_IMAGE_NAME=${DOCKER_MODULE_IMAGE_NAME} \
            VERRAZZANO_HELM_OPERATOR_IMAGE_NAME=${DOCKER_HELM_IMAGE_NAME} \
            VERRAZZANO_CALICO_OPERATOR_IMAGE_NAME=${DOCKER_CALICO_IMAGE_NAME} \
            CREATE_LATEST_TAG=${CREATE_LATEST_TAG}
        #${GO_REPO_PATH}/${GIT_REPO_DIR}/tools/scripts/generate_image_list.sh $WORKSPACE/generated-verrazzano-bom.json $WORKSPACE/verrazzano_images.txt
    """
}

// Called in Stage Generate operator.yaml steps
def generateOperatorYaml(dockerImageTag) {
    sh """
        case "${env.BRANCH_NAME}" in
            main|release-*)
                ;;
            *)
                echo "Adding image pull secrets to operator.yaml for non main/release branch"
                export IMAGE_PULL_SECRETS=verrazzano-container-registry
        esac

        echo "Generating operator manifests and versioned Charts"
        cd ${GO_REPO_PATH}/${GIT_REPO_DIR}
        make generate-operator-artifacts BUILD_DEPLOY=${WORKSPACE}/generated \
            DOCKER_REPO=${env.DOCKER_REPO} DOCKER_NAMESPACE=${env.DOCKER_NAMESPACE} DOCKER_IMAGE_TAG=${dockerImageTag} \
            VERRAZZANO_MODULE_OPERATOR_IMAGE_NAME=${DOCKER_MODULE_IMAGE_NAME} \
            VERRAZZANO_HELM_OPERATOR_IMAGE_NAME=${DOCKER_HELM_IMAGE_NAME} \
            VERRAZZANO_CALICO_OPERATOR_IMAGE_NAME=${DOCKER_CALICO_IMAGE_NAME}
    """
}

// Called in Stage Save Generated Files steps
def saveGeneratedFiles() {
    sh """
        cd ${GO_REPO_PATH}/verrazzano-modules
        #oci --region us-phoenix-1 os object put --force --namespace ${OCI_OS_NAMESPACE} -bn ${OCI_OS_BUCKET} --name ${env.BRANCH_NAME}/operator.yaml --file $WORKSPACE/generated-operator.yaml
        #oci --region us-phoenix-1 os object put --force --namespace ${OCI_OS_NAMESPACE} -bn ${OCI_OS_COMMIT_BUCKET} --name ephemeral/${env.BRANCH_NAME}/${SHORT_COMMIT_HASH}/operator.yaml --file $WORKSPACE/generated-operator.yaml
        #oci --region us-phoenix-1 os object put --force --namespace ${OCI_OS_NAMESPACE} -bn ${OCI_OS_BUCKET} --name ${env.BRANCH_NAME}/generated-verrazzano-bom.json --file $WORKSPACE/generated-verrazzano-bom.json
        #oci --region us-phoenix-1 os object put --force --namespace ${OCI_OS_NAMESPACE} -bn ${OCI_OS_COMMIT_BUCKET} --name ephemeral/${env.BRANCH_NAME}/${SHORT_COMMIT_HASH}/generated-verrazzano-bom.json --file $WORKSPACE/generated-verrazzano-bom.json
    """
}

// Called in Stage Clean workspace and checkout steps
@NonCPS
def getCommitList() {
    echo "Checking for change sets"
    def commitList = []
    def changeSets = currentBuild.changeSets
    for (int i = 0; i < changeSets.size(); i++) {
        echo "get commits from change set"
        def commits = changeSets[i].items
        for (int j = 0; j < commits.length; j++) {
            def commit = commits[j]
            def id = commit.commitId
            echo "Add commit id: ${id}"
            commitList.add(id)
        }
    }
    return commitList
}

def trimIfGithubNoreplyUser(userIn) {
    if (userIn == null) {
        echo "Not a github noreply user, not trimming: ${userIn}"
        return userIn
    }
    if (userIn.matches(".*\\+.*@users.noreply.github.com.*")) {
        def userOut = userIn.substring(userIn.indexOf("+") + 1, userIn.indexOf("@"))
        return userOut;
    }
    if (userIn.matches(".*<.*@users.noreply.github.com.*")) {
        def userOut = userIn.substring(userIn.indexOf("<") + 1, userIn.indexOf("@"))
        return userOut;
    }
    if (userIn.matches(".*@users.noreply.github.com")) {
        def userOut = userIn.substring(0, userIn.indexOf("@"))
        return userOut;
    }
    echo "Not a github noreply user, not trimming: ${userIn}"
    return userIn
}

def getSuspectList(commitList, userMappings) {
    def retValue = ""
    def suspectList = []
    if (commitList == null || commitList.size() == 0) {
        echo "No commits to form suspect list"
    } else {
        for (int i = 0; i < commitList.size(); i++) {
            def id = commitList[i]
            try {
                def gitAuthor = sh(
                    script: "git log --format='%ae' '$id^!'",
                    returnStdout: true
                ).trim()
                if (gitAuthor != null) {
                    def author = trimIfGithubNoreplyUser(gitAuthor)
                    echo "DEBUG: author: ${gitAuthor}, ${author}, id: ${id}"
                    if (userMappings.containsKey(author)) {
                        def slackUser = userMappings.get(author)
                        if (!suspectList.contains(slackUser)) {
                            echo "Added ${slackUser} as suspect"
                            retValue += " ${slackUser}"
                            suspectList.add(slackUser)
                        }
                    } else {
                        // If we don't have a name mapping use the commit.author, at least we can easily tell if the mapping gets dated
                        if (!suspectList.contains(author)) {
                            echo "Added ${author} as suspect"
                            retValue += " ${author}"
                            suspectList.add(author)
                        }
                    }
                } else {
                    echo "No author returned from git"
                }
            } catch (Exception e) {
                echo "INFO: Problem processing commit ${id}, skipping commit: " + e.toString()
            }
        }
    }
    def startedByUser = "";
    def causes = currentBuild.getBuildCauses()
    echo "causes: " + causes.toString()
    for (cause in causes) {
        def causeString = cause.toString()
        echo "current cause: " + causeString
        def causeInfo = readJSON text: causeString
        if (causeInfo.userId != null) {
            startedByUser = causeInfo.userId
        }
    }

    if (startedByUser.length() > 0) {
        echo "Build was started by a user, adding them to the suspect notification list: ${startedByUser}"
        def author = trimIfGithubNoreplyUser(startedByUser)
        echo "DEBUG: author: ${startedByUser}, ${author}"
        if (userMappings.containsKey(author)) {
            def slackUser = userMappings.get(author)
            if (!suspectList.contains(slackUser)) {
                echo "Added ${slackUser} as suspect"
                retValue += " ${slackUser}"
                suspectList.add(slackUser)
            }
        } else {
            // If we don't have a name mapping use the commit.author, at least we can easily tell if the mapping gets dated
            if (!suspectList.contains(author)) {
               echo "Added ${author} as suspect"
               retValue += " ${author}"
               suspectList.add(author)
            }
        }
    } else {
        echo "Build not started by a user, not adding to notification list"
    }
    echo "returning suspect list: ${retValue}"
    return retValue
}
