# Terraform AKS Demo Project

## Terraform - AKS - AKS GitOps (built-in Flux v2) - Azure Key Vault - Azure SQL database - GitHub Actions - DockerHub - Multi-container App

This repository contains a demo project showcasing the use of Terraform and Azure Kubernetes Service (AKS) to deploy a simple multi-container application on Azure using the GitOps approach.

## Overview

The project demonstrates how to use Terraform to provision Azure resources, including AKS, Key Vault, and Azure SQL Database. The infrastructure is designed to support a microservices-based application consisting of three services.

Service 1 (aks-demo-log): This service is accessible through a load balancer on a specific port. It serves as the entry point to the application.  
Service 2 (aks-demo-job): This service processes incoming requests and captures relevant statistics. It stores this data in the Azure SQL Database.  
Service 3 (aks-demo-bot): This service periodically sends the collected statistics to a Telegram bot for further analysis.  
The infrastructure also demonstrates how to use Managed Identity to enable secure access from AKS to Azure Key Vault secrets, ensuring a robust and secure deployment process.

By following the instructions in this repository, you will be able to deploy the entire microservices application along with the underlying infrastructure in a consistent and automated manner with a minimal number of steps.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) installed on your local machine.
- An Azure subscription with sufficient permissions to create resources.
- Get your telegram bot token  
  Create a bot from Telegram [@BotFather](https://t.me/BotFather) and obtain an access token.

## Project Structure

The repository is organized as follows:

- `main.tf`: Terraform configuration file defining the infrastructure resources.
- `variables.tf`: Defines variables used in the Terraform configuration.
- `terraform.tfvars`: Contains values for the variables defined in `variables.tf`. **NOTE**: Do not commit sensitive information such as passwords or access tokens in this file.
- `output.tf`: Defines the output values that are displayed after Terraform applies the configuration.
- `scripts/`: Contains any additional scripts used in the project.
- `.gitnub/workflows`: This directory contains the GitHub Actions workflows that automate the CI/CD process. These workflows handle tasks such as building and pushing Docker container images to DockerHub, and updateing Kubernetes manifests.  
- `log/`, `job/`, `bot/`: These directories contain the source code of the three microservices - log, job, and bot. Each directory includes a main.go file containing the application code.
- `clusters/flux-system/`: This directory holds the Kubernetes manifests and configurations for Flux, which is used to manage the GitOps workflow. Flux continuously monitors the Git repository for changes in the Kubernetes manifests and automatically updates the AKS cluster accordingly, ensuring that the desired state of the infrastructure and applications is always in sync with the Git repository.

## How to Use

1. Clone the repository to your local machine.
2. Create your own `terraform.tfvars` file with your desired values for the variables. **NOTE**: Make sure to add sensitive information, such as passwords, to your environment variables and reference them in the `terraform.tfvars` file.  
Example:  
```sh
aks-demo-kv-tg-token = "token"
aks-demo-sql-server-name     = "sql-server-name "
aks-demo-sql-server-login    = "server-login"
aks-demo-sql-server-password = "password"
aks-demo-sql-server-dbname   = "dbname"
```
3. Create a new public GitHub repository and add the following two repository secrets:  
`DOCKERHUB_USERNAME`: Your username on DockerHub  
`DOCKERHUB_TOKEN`: Your DockerHub token
4. Make any commits to the main.go files of the log, job, and bot apps to initiate the generation of container images.
After the images are created, the name of the new container will be automatically added as a commit to the Kubernetes manifests.  
These manifests will then be processed by Flux on the AKS side, ensuring the implementation of CI/CD using the GitOps approach provided here.  
After the images are created, ensure that the Docker repository is public.  
Note: It's possible to use any other container registry, but in that case, the workflows will need to be adjusted accordingly.
5. Initialize Terraform by running the following command:
```sh
terraform init
```
6. Review the changes that Terraform will make by running the following command:
```sh
terraform plan -var-file=terraform.tfvars
```
7. If the plan looks correct, apply the configuration with the following command:
```sh
terraform apply -var-file=terraform.tfvars
```
8. After the deployment is complete, Terraform will output the necessary information to access the deployed resources.

## Deploy Additional Manifests

To deploy additional manifests, run the following commands:
```sh
chmod +x scripts/deploy_additional_manifests.sh
./scripts/deploy_additional_manifests.sh
```
! Note: Make sure to take note of the IP address from the script output.

## Test application

This project provides a test application that demonstrates the functionality of the microservices deployed on AKS (Azure Kubernetes Service). After setting up the infrastructure using Terraform, you can verify the successful deployment of the application by following these steps:

Check Application Deployment:  
Run the following commands to ensure that the application has been successfully deployed on AKS:
```sh
kubectl get deploy -n aks-demo
NAME           READY   UP-TO-DATE   AVAILABLE   AGE
tf-aks-demo-bot   1/1     1            1           3m43s
tf-aks-demo-job   1/1     1            1           26h
tf-aks-demo-log   2/2     2            2           26h

kubectl get pod -n aks-demo
NAME                            READY   STATUS    RESTARTS        AGE
tf-aks-demo-bot-7469f846b5-7q8gg   1/1     Running   0               3m50s
tf-aks-demo-job-7f4b867c99-2szvb   1/1     Running   0               11h
tf-aks-demo-log-6ff8f6754f-29vsx   1/1     Running   0               3h54m
tf-aks-demo-log-6ff8f6754f-fsfbq   1/1     Running   0               3h54m
```

Test the Load Balancer:  
The first microservice can be accessed through the load balancer on a specific port. Use the following command multiple times to test the load balancer's behavior:
```sh
curl <IP address from the previous step>:8081
```
Example Output:
```sh
curl 52.151.211.174:8081
UID: 3592-9764-4881-2311
Request Time: 2023-07-21T12:00:00Z%

curl 52.151.211.174:8081
UID: 3592-9764-4881-2311
Request Time: 2023-07-21T12:00:01Z%

curl 52.151.211.174:8081
UID: 1122-6499-5913-323
Request Time: 2023-07-21T12:00:02Z%
```

Here, you'll observe that the load balancer distributes traffic between the two instances of the first microservice. Each instance has its unique UID visible in the response.

Statistics Aggregation:  
The second microservice is responsible for selecting and aggregating statistics about requests per hour. It stores this information in a separate table in the Azure SQL Database.

Start the Telegram Bot:  
To enable the third microservice to send statistics to the Telegram bot, you need to start the bot by sending a message (e.g., /start) in the chat. After sending the initial message, the bot will begin its operation and regularly send statistics updates to the Telegram chat.

Note: If the bot encounters any issues in determining the chat_id during startup, it will wait for the first message in the chat to obtain the chat_id automatically.


## Cleanup

To destroy all the resources created by Terraform and avoid unnecessary costs, run the following command:

terraform destroy -var-file=terraform.tfvars

## Contributions

Contributions to this demo project are welcome! If you find any issues or have suggestions for improvements, please feel free to create a pull request.

## License

This project is licensed under the [MIT License](LICENSE).