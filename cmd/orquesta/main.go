package main

import (
	"app-platform/internals/aws"
	"fmt"
)

func main() {

	region := "us-east-1"
	//currentInstance := "i-0d4f73312571601fa"
	imageID := "ami-0ca6dfcefdde35d25"
	instanceType := "t2.micro"
	launchConfigName := "app-platform"
	cidr := "172.31.0.0/20"

	ob := &aws.OrchestratorBuilder{}
	awsServiceMangment := ob.
		Region(region).
		EnableEc2().
		AdquireVpc().
		CreateSubnet(cidr).
		CreateSecurityGroupConfiguration("app-platform", "app platform security group", ob.Orchestrator.VpcID).
		InputSecurityRule(ob.Orchestrator.SecurityGroup.GroupName).
		CreateLaunchConfiguration(imageID, instanceType, launchConfigName).
		Build()

	fmt.Println(awsServiceMangment)
}

func mains() {

	//ami := flag.String("ami", "", "You need to set the AMI to start the deployment")
	//vpc := flag.String("vpc", "", "VPC to load")
	//securityGroupIDLoaded := flag.String("sg", "", "Security Group to load")
	//var securityGroupID, vpcID string
	//flag.Parse()

	/*
		if *ami == "" {
			flag.PrintDefaults()
			os.Exit(1)
		}*/

	/*
		awsOS := (*aws.OrchestratorService).New("us-east-1")

		awsOS.AdquireVpc()

		fmt.Println(awsOS)
		fmt.Printf("ami id: %s\n", *ami)
	*/
}
