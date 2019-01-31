package main

import (
	"app-platform/internals/aws"
	"fmt"
)

func main() {

	region := "us-east-1"
	imageID := "ami-0ca6dfcefdde35d25"
	instanceType := "t2.micro"
	launchConfigName := "app-platform"
	cidr := "172.31.0.0/20"
	elbName := "app-platform"

	aws.PrintMessage("\n---------------------------------------")
	aws.PrintMessage("Orquesta App has been started.........")

	ob := &aws.OrchestratorBuilder{}
	awsServiceMangment := ob.
		NewSession(region).
		EnableEc2().
		CreateVpc().
		CreateSubnet(cidr).
		CreateLoadBalancer(elbName, 80, 80).
		CreateSecurityGroupConfiguration("app-platform", "app platform security group", ob.Orchestrator.VpcID).
		InputSecurityRule(ob.Orchestrator.SecurityGroup.GroupName).
		CreateLaunchConfiguration(imageID, instanceType, launchConfigName).
		DeployAutoScalingGroup("app-platform", 1, 2).
		Build()

	fmt.Println("\n------- summary --------")
	fmt.Printf("%+v\n", awsServiceMangment)

}
