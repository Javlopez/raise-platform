package aws

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//SecurityGroup struct
type SecurityGroup struct {
	GroupID string
}

//OrchestratorManagement struct
type OrchestratorManagement struct {
	SecurityGroup
	Region, VpcID string
	ec2           *ec2.EC2
}

//OrchestratorBuilder service
type OrchestratorBuilder struct {
	OrchestratorManagement
}

//Build method
func (ob *OrchestratorBuilder) Build() OrchestratorManagement {
	return ob.OrchestratorManagement
}

//Region method
func (ob *OrchestratorBuilder) Region(region string) *OrchestratorBuilder {
	ob.OrchestratorManagement.Region = region
	return ob
}

//EnableEc2 method
func (ob *OrchestratorBuilder) EnableEc2() *OrchestratorBuilder {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Region: aws.String(ob.OrchestratorManagement.Region)},
	}))

	ec2 := ec2.New(sess)
	ob.OrchestratorManagement.ec2 = ec2
	return ob
}

//AdquireVpc method
func (ob *OrchestratorBuilder) AdquireVpc() *OrchestratorBuilder {
	result, _ := ob.OrchestratorManagement.ec2.DescribeVpcs(nil)
	if len(result.Vpcs) == 0 {
		fmt.Fprintf(os.Stderr, "No Vpc found")
		os.Exit(1)
	}
	ob.OrchestratorManagement.VpcID = aws.StringValue(result.Vpcs[0].VpcId)
	return ob
}

//CreateSecurityGroupConfiguration method
func (ob *OrchestratorBuilder) CreateSecurityGroupConfiguration(name string, description string, vpcID string) *OrchestratorBuilder {
	securityGroup, err := ob.OrchestratorManagement.ec2.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(name),
		Description: aws.String(description),
		VpcId:       aws.String(vpcID),
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create security group: %s %v", name, err)
		os.Exit(1)
	}

	ob.OrchestratorManagement.SecurityGroup.GroupID = *securityGroup.GroupId
	return ob
}

func printError(message string, e error) {
	fmt.Println(message, e.Error())
	log.Fatal(e.Error())
}
