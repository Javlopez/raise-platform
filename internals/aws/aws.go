package aws

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
	Name    string
	MaxSize int64
	MinSize int64
}

//LoadBalancer struct
type LoadBalancer struct {
	Name string
}

//InstancesCollection struct
type InstancesCollection struct {
	Instances []*ec2.Reservation
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
	session *session.Session
	InstancesCollection
}

//OrchestratorBuilder service
type OrchestratorBuilder struct {
	Orchestrator
}

//Build method
func (ob *OrchestratorBuilder) Build() Orchestrator {
	return ob.Orchestrator

}

//NewSession method
func (ob *OrchestratorBuilder) NewSession(region string) *OrchestratorBuilder {
	ob.Orchestrator.Region = region
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Region: aws.String(ob.Orchestrator.Region)},
	}))

	PrintMessage("Starting new session with Amazon AWS Region " + ob.Orchestrator.Region)
	ob.Orchestrator.session = sess
	return ob
}

//EnableEc2 method
func (ob *OrchestratorBuilder) EnableEc2() *OrchestratorBuilder {

	ec2 := ec2.New(ob.Orchestrator.session)
	ob.Orchestrator.ec2 = ec2
	PrintMessage("Enabling EC2 service")
	return ob
}

//CreateVpc method
func (ob *OrchestratorBuilder) CreateVpc() *OrchestratorBuilder {
	result, _ := ob.Orchestrator.ec2.DescribeVpcs(nil)
	if len(result.Vpcs) == 0 {
		fmt.Fprintf(os.Stderr, "No Vpc found")
		os.Exit(1)
	}

	PrintMessage("Loading VPC ID: " + *result.Vpcs[0].VpcId)
	ob.Orchestrator.VpcID = aws.StringValue(result.Vpcs[0].VpcId)
	return ob
}

//CreateSubnet method
func (ob *OrchestratorBuilder) CreateSubnet(cidr string) *OrchestratorBuilder {

	if ob.FindSubnet() {
		message := "Subnet " + ob.Orchestrator.Subnet.ID + " CIDR: " + cidr + " was created in VPC: " + ob.Orchestrator.VpcID
		PrintMessage(message)
		return ob
	}

	inputSubnet := &ec2.CreateSubnetInput{
		CidrBlock: aws.String(cidr),
		VpcId:     aws.String(ob.Orchestrator.VpcID),
	}

	result, err := ob.Orchestrator.ec2.CreateSubnet(inputSubnet)
	if err != nil {

		aerr, _ := err.(awserr.Error)

		if aerr.Code() == "InvalidSubnet.Conflict" {
			message := "Subnet " + ob.Orchestrator.Subnet.ID + " CIDR: " + cidr + " was loaded in VPC: " + ob.Orchestrator.VpcID
			PrintMessage(message)
			return ob
		}
		errMessage := "Cannot to create a subnet with CIDR: " + cidr
		printError(errMessage, err)
		return nil
	}

	ob.Orchestrator.Subnet.ID = *result.Subnet.SubnetId
	message := "Subnet " + ob.Orchestrator.Subnet.ID + "CIDR: " + cidr + " was created in VPC: " + ob.Orchestrator.VpcID
	PrintMessage(message)

	return ob
}

//FindSubnet method
func (ob *OrchestratorBuilder) FindSubnet() bool {

	params := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(ob.Orchestrator.VpcID),
				},
			},
		},
	}

	result, err := ob.ec2.DescribeSubnets(params)
	if err != nil {
		printError("We cannot describe the subnets", err)
	}

	if len(result.Subnets) == 0 {
		return false
	}

	subnet := result.Subnets[0]
	ob.Orchestrator.Subnet.ID = *subnet.SubnetId
	return true
}

//CreateSecurityGroupConfiguration method
func (ob *OrchestratorBuilder) CreateSecurityGroupConfiguration(name string, description string, vpcID string) *OrchestratorBuilder {

	ob.FindSecurityGroup(name)

	if len(ob.Orchestrator.SecurityGroup.GroupID) > 0 {
		message := "The security group:" + name + " with ID " + ob.Orchestrator.SecurityGroup.GroupID + " was loaded"
		PrintMessage(message)
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

	message := "The security group:" + name + " ID " + ob.Orchestrator.SecurityGroup.GroupID + " was created"
	PrintMessage(message)
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
				PrintMessage("The security Rule was already added into security group:" + ob.Orchestrator.SecurityGroup.GroupID)
				return ob
			}

			errMessage := "Cannot add security rules in: " + groupName
			printError(errMessage, err)
		}
	}

	PrintMessage("The security rule was added into security group" + ob.Orchestrator.SecurityGroup.GroupID)
	return ob
}

//CreateLoadBalancer CreateLoadBalancer
func (ob *OrchestratorBuilder) CreateLoadBalancer(elbName string, instancePort int64, loadBalancerPort int64) *OrchestratorBuilder {

	elbService := elb.New(ob.Orchestrator.session)
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

	ob.Orchestrator.LoadBalancer.Name = elbName

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elb.ErrCodeDuplicateAccessPointNameException:
				PrintMessage("Loading ELB: " + elbName)
				return ob
			default:
				message := "Error creating ELB : " + elbName
				printError(message, err)
			}
		}
		message := "Error creating ELB : " + elbName
		printError(message, err)
	}

	PrintMessage("ELB: " + elbName + " created")
	return ob
}

//CreateLaunchConfiguration method
func (ob *OrchestratorBuilder) CreateLaunchConfiguration(imageID string, instanceType string, launchConfigurationName string) *OrchestratorBuilder {

	autoScalingAws := autoscaling.New(ob.Orchestrator.session)

	inputParams := &autoscaling.CreateLaunchConfigurationInput{
		ImageId:                 aws.String(imageID),
		InstanceType:            aws.String(instanceType),
		LaunchConfigurationName: aws.String(launchConfigurationName),
		SecurityGroups: []*string{
			aws.String(ob.Orchestrator.SecurityGroup.GroupID),
		},
	}

	_, err := autoScalingAws.CreateLaunchConfiguration(inputParams)

	ob.Orchestrator.LaunchConfiguration.Name = launchConfigurationName

	if err != nil {

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case autoscaling.ErrCodeAlreadyExistsFault:
					PrintMessage("Load Launch configuration:" + launchConfigurationName)
					return ob
				default:
					errMessage := "Cannot be create the autoscaling group:" + launchConfigurationName
					printError(errMessage, err)
				}
			}
		}

		errMessage := "Error creating launch configuration: " + launchConfigurationName
		printError(errMessage, err)
	}

	PrintMessage("Launch Configuration:" + launchConfigurationName + " Created ")
	return ob
}

// DeployAutoScalingGroup method
func (ob *OrchestratorBuilder) DeployAutoScalingGroup(name string, minSize int64, maxSize int64) *OrchestratorBuilder {

	autoScalingAws := autoscaling.New(ob.Orchestrator.session)
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

	_, err := autoScalingAws.CreateAutoScalingGroup(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case autoscaling.ErrCodeAlreadyExistsFault:

				inputUpdate := &autoscaling.UpdateAutoScalingGroupInput{
					AutoScalingGroupName:    aws.String(name),
					LaunchConfigurationName: aws.String(ob.Orchestrator.LaunchConfiguration.Name),
					MaxSize:                 aws.Int64(maxSize),
					MinSize:                 aws.Int64(minSize),
					VPCZoneIdentifier:       aws.String(ob.Orchestrator.Subnet.ID),
					HealthCheckGracePeriod:  aws.Int64(120),
					HealthCheckType:         aws.String("ELB"),
				}

				_, err := autoScalingAws.UpdateAutoScalingGroup(inputUpdate)

				if err != nil {
					message := "Cannot be update the autoscaling group:" + name
					printError(message, err)
				}
				ob.Orchestrator.AutoscalingGroup.MaxSize = maxSize
				ob.Orchestrator.AutoscalingGroup.MinSize = minSize
				ob.Orchestrator.AutoscalingGroup.Name = name

				PrintMessage("AutoScaling: " + name + " has been deployed ")
				return ob
			default:
				message := "Cannot be create the autoscaling group:" + name
				printError(message, err)
			}
		}
	}

	ob.Orchestrator.AutoscalingGroup.MaxSize = maxSize
	ob.Orchestrator.AutoscalingGroup.MinSize = minSize
	ob.Orchestrator.AutoscalingGroup.Name = name
	PrintMessage("AutoScaling: " + name + " has been deployed ")
	return ob
}

//FindInstancesByAMI method
func (ob *OrchestratorBuilder) FindInstancesByAMI(ami string) *OrchestratorBuilder {

	PrintMessage("Looking instances with AMI: " + ami)
	inputParams := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("image-id"),
				Values: []*string{
					aws.String(ami),
				},
			},
		},
	}

	result, err := ob.Orchestrator.ec2.DescribeInstances(inputParams)

	if err != nil {
		fmt.Println(err.Error())
	}

	instances := strconv.Itoa(len(result.Reservations))
	PrintMessage(instances + " instances were founded....")

	ob.Orchestrator.InstancesCollection.Instances = result.Reservations

	return ob
}

//PrintMessage func
func PrintMessage(message string) {
	fmt.Printf("%v\n", message)
}

func printError(message string, e error) {
	fmt.Println(message, e.Error())
	log.Fatal(e.Error())
}
