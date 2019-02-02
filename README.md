# APP Platform

### Description
This software contains two applications to achive work better with your infrastucture, especifically Amazon AWS

#### Requeriments
- Docker 

#### Applications
- orquesta:
    - This application is like a micro-terraform app and allow to setup an autoscalable architecture, creating a VPC, a subnet, a launch configuration, an autoscaling group, an security group, and ELB with some specified parameters
- deploy 
    - Deploy application allow us to change the launch configuration to another launch configuration using another AMI


## Instructions to install the app

1. Clone Application
```
git clone git@github.com:Javlopez/raise-platform.git 
```    

2.- Build the container
```
$ cd raise-platform
$ docker build . -t app-platform --no-cache --build-arg AWS_ACCESS_KEY_ID=AKIAJNPGNIGDVNJTX2EA --build-arg  AWS_SECRET_ACCESS_KEY=xxxxxxxx
```

3.- Run the container

```
docker run --entrypoint /bin/bash -it app-platform 
```

4.- Run __orquesta__, this app was created to deploy an AMI inside a VPC using an autoscaling architecture this includes:

 - Create a VPC or choose one
 - Create only one subnet
 - Create a Load Balancer
 - Create Security Group
    -  Attach security Rule (allowing communication by port 80)
 - Create Launch configuration
 - Create AutoScaling Group
 ```
  $ orquesta --ami=ami-0ca6dfcefdde35d25 --vpc=vpc-d69537ac --sg=sg-009e0b8799cbad4a7
 ```

### Parameters, all of them are optionals

 - --deploy=plan
    - This is the type of execution thst will be execute
        - plan: Just show a summary of the changes
        - execute: The plan will be executed

 - --ami=ami-0ca6dfcefdde35d25
    - AMI id required to build the launch configuration

 - --region=us-east-1
    - Region where your infrastructure will be deployed

 - --vpc=vpc-d69537ac
    - ID of the VPC

 - --instance-type=t2.micro
    - Type of the instance to be deployed

 - --elb=elb-name
    - Elastic Load Balancer name

 - --security-group-name=app-platform
    - Security Group Name

 - --launch-config-name=lanunch-name
    - Launch configuration name

 - --cidr=172.31.0.0/20
    - CIDR for the subnet

 - --autoscaling-group-name=app-platform
    - Name of AutoScaling Group Name

-  --elb-instance-port=80
    - Port of the instance 

-  --elb-load-balancer-port=80
    - Port of the load balancer  