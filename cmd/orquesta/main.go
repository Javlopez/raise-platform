package main

import (
	"app-platform/internals/aws"
	"flag"
	"fmt"
	"strconv"
)

func main() {

	ami := flag.String("ami", "ami-0ca6dfcefdde35d25", "You need to set the AMI to start the deployment")
	region := flag.String("region", "us-east-1", "Region")
	vpc := flag.String("vpc", "vpc-d69537ac", "Vpc id")
	instanceType := flag.String("instance-type", "t2.micro", "Instance Type")
	elbName := flag.String("elb", "app-platform", "Elastic Load Balancer")
	sgName := flag.String("security-group-name", "app-platform", "Elastic Load Balancer")
	launchConfigName := flag.String("launch-config-name", "app-platform", "Launch Config Name")
	cidr := flag.String("cidr", "172.31.0.0/20", "CIDR block")
	autoScalingName := flag.String("autoscaling-group-name", "app-platform", "Name of the Autoscaling Group")
	deploy := flag.String("deploy", "plan", "Type of the execution")
	instancePort := flag.Int64("elb-instance-port", 80, "Instance PORT")
	elbPort := flag.Int64("elb-load-balancer-port", 80, "Load balancer PORT")
	flag.Parse()

	const MIN_INSTANCES_PER_GROUP = 2
	const MAX_INSTANCES_PER_GROUP = 4

	aws.PrintMessage("\n---------------------------------------")
	aws.PrintMessage("Orquesta App has been started.........")
	aws.PrintMessage("Building blocks using AMI:" + *ami)

	if *deploy == "plan" {
		aws.PrintMessage("-------- MODE: TEST -----------------")
		aws.PrintMessage("A VPC will be deployed VPCID:" + *vpc)
		aws.PrintMessage("Region to use:" + *region)
		aws.PrintMessage("Instance Type:" + *instanceType)
		aws.PrintMessage("Load Balancer Name:" + *elbName)
		aws.PrintMessage("Load balancer PORT:" + strconv.Itoa(int(*elbPort)))
		aws.PrintMessage("Instance PORT:" + strconv.Itoa(int(*instancePort)))
		aws.PrintMessage("Security Group Name:" + *sgName)
		aws.PrintMessage("Launch Configuration Name:" + *launchConfigName)
		aws.PrintMessage("CIDR Block:" + *cidr)
		aws.PrintMessage("Autoscaling Group Name:" + *autoScalingName)
		return
	}

	ob := &aws.OrchestratorBuilder{}
	awsServiceMangment := ob.
		NewSession(*region).
		EnableEc2().
		CreateVpc(*vpc).
		CreateSubnet(*cidr).
		CreateLoadBalancer(*elbName, *instancePort, *elbPort).
		CreateSecurityGroupConfiguration(*sgName, "app platform security group", ob.Orchestrator.VpcID).
		InputSecurityRule(ob.Orchestrator.SecurityGroup.GroupName).
		CreateLaunchConfiguration(*ami, *instanceType, *launchConfigName).
		DeployAutoScalingGroup(*autoScalingName, MIN_INSTANCES_PER_GROUP, MAX_INSTANCES_PER_GROUP).
		Build()

	fmt.Println("\n---------- summary ----------")
	fmt.Printf("%+v\n", awsServiceMangment)
	fmt.Println("\n-----------------------------")

}
