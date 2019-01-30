package aws

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
)

//SecurityGroup struct
type SecurityGroup struct {
	GroupID   string
	GroupName string
}

//LaunchConfiguration struct
type LaunchConfiguration struct {
	Name string
}

//Subnet struct
type Subnet struct {
	ID string
}

//AutoscalingGroup struct
type AutoscalingGroup struct {
	Name string
}

//LoadBalancer struct
type LoadBalancer struct {
	Name string
}

// Orchestrator struct
type Orchestrator struct {
	Subnet
	SecurityGroup
	Region, VpcID string
	ec2           *ec2.EC2
	LaunchConfiguration
	AutoscalingGroup
	LoadBalancer
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

//CreateSubnet method
func (ob *OrchestratorBuilder) CreateSubnet(cidr string) *OrchestratorBuilder {
	inputSubnet := &ec2.CreateSubnetInput{
		CidrBlock: aws.String(cidr),
		VpcId:     aws.String(ob.Orchestrator.VpcID),
	}

	result, err := ob.Orchestrator.ec2.CreateSubnet(inputSubnet)
	if err != nil {

		aerr, _ := err.(awserr.Error)

		if aerr.Code() == "InvalidSubnet.Conflict" {
			return ob
		}
		message := "Cannot to create a subnet with CIDR: " + cidr
		printError(message, err)
		return nil
	}

	ob.Orchestrator.Subnet.ID = *result.Subnet.SubnetId

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
	ob.Orchestrator.SecurityGroup.GroupName = name
	return ob
}

//FindSecurityGroup method
func (ob *OrchestratorBuilder) FindSecurityGroup(securityGroupName string) *OrchestratorBuilder {
	response, err := ob.Orchestrator.ec2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("group-name"),
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
	ob.Orchestrator.SecurityGroup.GroupName = securityGroupName
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
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "InvalidPermission.Duplicate" {
				return ob
			}

			message := "Cannot add security rules in: " + groupName
			printError(message, err)
		}
	}
	return ob
}

//CreateLoadBalancer CreateLoadBalancer
func (ob *OrchestratorBuilder) CreateLoadBalancer(elbName string, instancePort int64, loadBalancerPort int64) *OrchestratorBuilder {

	elbService := elb.New(session.New())
	input := &elb.CreateLoadBalancerInput{
		Listeners: []*elb.Listener{
			{
				InstancePort:     aws.Int64(instancePort),
				InstanceProtocol: aws.String("HTTP"),
				LoadBalancerPort: aws.Int64(loadBalancerPort),
				Protocol:         aws.String("HTTP"),
			},
		},
		LoadBalancerName: aws.String(elbName),
		SecurityGroups: []*string{
			aws.String(ob.Orchestrator.SecurityGroup.GroupID),
		},
		Subnets: []*string{
			aws.String(ob.Orchestrator.Subnet.ID),
		},
	}

	_, err := elbService.CreateLoadBalancer(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elb.ErrCodeDuplicateAccessPointNameException:
			default:
				message := "Error creating ELB : " + elbName
				printError(message, err)
			}
		}
		message := "Error creating ELB : " + elbName
		printError(message, err)
	}

	ob.Orchestrator.LoadBalancer.Name = elbName

	return ob
}

//CreateLaunchConfiguration method
func (ob *OrchestratorBuilder) CreateLaunchConfiguration(imageID string, instanceType string, launchConfigurationName string) *OrchestratorBuilder {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Region: aws.String(ob.Orchestrator.Region)},
	}))

	autoScalingAws := autoscaling.New(sess)

	inputParams := &autoscaling.CreateLaunchConfigurationInput{
		ImageId:                 aws.String(imageID),
		InstanceType:            aws.String(instanceType),
		LaunchConfigurationName: aws.String(launchConfigurationName),
		SecurityGroups: []*string{
			aws.String(ob.Orchestrator.SecurityGroup.GroupID),
		},
	}

	_, err := autoScalingAws.CreateLaunchConfiguration(inputParams)

	if err != nil {

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case autoscaling.ErrCodeAlreadyExistsFault:
					return ob
				default:
					message := "Cannot be create the autoscaling group:" + launchConfigurationName
					printError(message, err)
				}
			}
		}

		message := "Error creating launch configuration: " + launchConfigurationName
		printError(message, err)
	}

	ob.Orchestrator.LaunchConfiguration.Name = launchConfigurationName
	return ob

}

// CreateAutoScalingGroup method
func (ob *OrchestratorBuilder) CreateAutoScalingGroup(name string, minSize int64, maxSize int64) *OrchestratorBuilder {

	autoScalingAws := autoscaling.New(session.New())
	input := &autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName:    aws.String(name),
		LaunchConfigurationName: aws.String(ob.Orchestrator.LaunchConfiguration.Name),
		MaxSize:                 aws.Int64(maxSize),
		MinSize:                 aws.Int64(minSize),
		VPCZoneIdentifier:       aws.String(ob.Orchestrator.Subnet.ID),
		HealthCheckGracePeriod:  aws.Int64(120),
		HealthCheckType:         aws.String("ELB"),
		LoadBalancerNames: []*string{
			aws.String(ob.Orchestrator.LoadBalancer.Name),
		},
	}

	result, err := autoScalingAws.CreateAutoScalingGroup(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case autoscaling.ErrCodeAlreadyExistsFault:
				return ob
			default:
				message := "Cannot be create the autoscaling group:" + name
				printError(message, err)
			}
		}
	}

	fmt.Println(result)
	return nil

	return ob
}

func printError(message string, e error) {
	fmt.Println(message, e.Error())
	log.Fatal(e.Error())
}
