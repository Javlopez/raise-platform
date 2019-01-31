package main

import (
	"app-platform/internals/aws"
	"flag"
	"fmt"
	"os"
)

func main() {

	ami := flag.String("ami", "ami-0ca6dfcefdde35d25", "You need to set the AMI to start the deployment")
	region := flag.String("region", "us-east-1", "Region")
	instanceType := flag.String("instance-type", "t2.micro", "Instance Type")
	elbName := flag.String("elb", "app-platform", "Elastic Load Balancer")
	launchConfigName := flag.String("launch-config-name", "app-platform", "Launch Config Name")
	cidr := flag.String("cidr", "172.31.0.0/20", "CIDR block")
	flag.Parse()

	const MIN_INSTANCES_PER_GROUP = 2
	const MAX_INSTANCES_PER_GROUP = 4

	if *ami == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	aws.PrintMessage("\n---------------------------------------")
	aws.PrintMessage("Orquesta App has been started.........")
	aws.PrintMessage("Building blocks using AMI:" + *ami)

	ob := &aws.OrchestratorBuilder{}
	awsServiceMangment := ob.
		NewSession(*region).
		EnableEc2().
		CreateVpc().
		CreateSubnet(*cidr).
		CreateLoadBalancer(*elbName, 80, 80).
		CreateSecurityGroupConfiguration("app-platform", "app platform security group", ob.Orchestrator.VpcID).
		InputSecurityRule(ob.Orchestrator.SecurityGroup.GroupName).
		CreateLaunchConfiguration(*ami, *instanceType, *launchConfigName).
		DeployAutoScalingGroup("app-platform", MIN_INSTANCES_PER_GROUP, MAX_INSTANCES_PER_GROUP).
		Build()

	fmt.Println("\n---------- summary ----------")
	fmt.Printf("%+v\n", awsServiceMangment)
	fmt.Println("\n-----------------------------")

}
