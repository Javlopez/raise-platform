package aws

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//SecurityGroup struct
type SecurityGroup struct {
	GroupID string
}

// Orchestrator struct
type Orchestrator struct {
	SecurityGroup
	Region, VpcID string
	ec2           *ec2.EC2
}

//OrchestratorBuilder service
type OrchestratorBuilder struct {
	Orchestrator
}

//Build method
func (ob *OrchestratorBuilder) Build() Orchestrator {
	return ob.Orchestrator

}

//Region method
func (ob *OrchestratorBuilder) Region(region string) *OrchestratorBuilder {
	ob.Orchestrator.Region = region
	return ob
}

//EnableEc2 method
func (ob *OrchestratorBuilder) EnableEc2() *OrchestratorBuilder {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Region: aws.String(ob.Orchestrator.Region)},
	}))

	ec2 := ec2.New(sess)
	ob.Orchestrator.ec2 = ec2
	return ob
}

//AdquireVpc method
func (ob *OrchestratorBuilder) AdquireVpc() *OrchestratorBuilder {
	result, _ := ob.Orchestrator.ec2.DescribeVpcs(nil)
	if len(result.Vpcs) == 0 {
		fmt.Fprintf(os.Stderr, "No Vpc found")
		os.Exit(1)
	}
	ob.Orchestrator.VpcID = aws.StringValue(result.Vpcs[0].VpcId)
	return ob
}

//CreateSecurityGroupConfiguration method
func (ob *OrchestratorBuilder) CreateSecurityGroupConfiguration(name string, description string, vpcID string) *OrchestratorBuilder {

	ob.FindSecurityGroup(name)

	if len(ob.Orchestrator.SecurityGroup.GroupID) > 0 {
		return ob
	}
	securityGroup, err := ob.Orchestrator.ec2.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(name),
		Description: aws.String(description),
		VpcId:       aws.String(vpcID),
	})

	if err != nil {
		message := "Unable to create security group: " + name
		printError(message, err)
		os.Exit(1)
	}

	ob.Orchestrator.SecurityGroup.GroupID = *securityGroup.GroupId
	return ob
}

//FindSecurityGroup method
func (ob *OrchestratorBuilder) FindSecurityGroup(securityGroupName string) *OrchestratorBuilder {
	response, err := ob.Orchestrator.ec2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("name"),
				Values: []*string{
					aws.String(securityGroupName),
				},
			},
		},
	})

	if err != nil {
		message := "error describing SecurityGroup" + securityGroupName
		printError(message, err)
		return nil
	}

	if len(response.SecurityGroups) == 0 {
		return nil
	}

	securityGroup := response.SecurityGroups[0]
	ob.Orchestrator.SecurityGroup.GroupID = *securityGroup.GroupId

	return ob
}

//InputSecurityRule method
func (ob *OrchestratorBuilder) InputSecurityRule(groupName string) *OrchestratorBuilder {
	_, err := ob.Orchestrator.ec2.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupName: aws.String(groupName),
		IpPermissions: []*ec2.IpPermission{
			(&ec2.IpPermission{}).
				SetIpProtocol("tcp").
				SetFromPort(80).
				SetToPort(80).
				SetIpRanges([]*ec2.IpRange{
					{CidrIp: aws.String("0.0.0.0/0")},
				}),
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot add security rules in: %s  error:%v", groupName, err)
	}
	return ob
}

func createLaunchConfiguration(awsAutoScalingService *autoscaling.AutoScaling) {
	fmt.Println("something")
}

func printError(message string, e error) {
	fmt.Println(message, e.Error())
	log.Fatal(e.Error())
}
