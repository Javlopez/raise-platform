package main

import (
	awsInternal "app-platform/internals/aws"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//InstanceEc2 struct
type InstanceEc2 struct {
	InstanceType string
}

//AWS_AUTOSCALING_GROUP_NAME const
const AWS_AUTOSCALING_GROUP_NAME = "aws:autoscaling:groupName"

func main() {

	if len(os.Args) < 3 {
		log.Fatal("Error: you need to provide old ami and new ami: deploy ami-123 ami-456")
		os.Exit(1)
	}

	oldAmi := os.Args[1]
	newAmi := os.Args[2]
	region := "us-east-1"

	awsInternal.PrintMessage("\n---------------------------------------")
	awsInternal.PrintMessage("Deploy App has been started.........")

	ob := &awsInternal.OrchestratorBuilder{}
	ob.
		NewSession(region).
		EnableEc2().
		FindInstancesByAMI(oldAmi).
		Build()

	instances := ob.Orchestrator.Instances

	if len(instances) == 0 {
		awsInternal.PrintMessage("O instances found, we cannot continue without instances")
		os.Exit(1)
	}

	awsInternal.PrintMessage("Replacing with AMI: " + newAmi)

	instance := *instances[0].Instances[0]

	autoScalingGroupName := findByAutoScalingTag(instance.Tags)
	vpcID := *instance.VpcId

	autoScalingGroups := awsInternal.GetAutoScalingGroups(ob, autoScalingGroupName)

	autoScalingGroup := GetAutoScalingGroup(autoScalingGroups)

	launchConfigurationName := *autoScalingGroup.LaunchConfigurationName

	launchConfigurations := awsInternal.LoadLaunchConfig(ob, launchConfigurationName)

	launchConfiguration := launchConfigurations.LaunchConfigurations[0]

	ob.
		CreateVpc(vpcID).
		CreateSecurityGroupConfiguration("app-platform", "app platform security group", ob.Orchestrator.VpcID).
		CreateLaunchConfiguration(newAmi, *launchConfiguration.InstanceType, "app2-platform")

	awsInternal.UpdateAutoScalingGroup(ob, autoScalingGroupName, launchConfigurationName)

	awsInternal.PrintMessage("Replace done ")
	awsInternal.PrintMessage("-------------------- summary -------------")
	autoScalingGroupsUpdated := awsInternal.GetAutoScalingGroups(ob, autoScalingGroupName)
	autoScalingGroupUpdated := GetAutoScalingGroup(autoScalingGroupsUpdated)

	fmt.Println(autoScalingGroupUpdated)

	awsInternal.PrintMessage("------------------------------------------")

}

func findByAutoScalingTag(tags []*ec2.Tag) string {
	for i := 0; i < len(tags); i++ {
		tag := tags[i]

		if *tag.Key == AWS_AUTOSCALING_GROUP_NAME {
			return *tag.Value
		}
	}
	return ""
}

//GetAutoScalingGroup func
func GetAutoScalingGroup(asg *autoscaling.DescribeAutoScalingGroupsOutput) *autoscaling.Group {
	return asg.AutoScalingGroups[0]
}
