package main

import (
	awsInternal "app-platform/internals/aws"
	"fmt"
	"log"
	"os"
)

//InstanceEc2 struct
type InstanceEc2 struct {
	InstanceType string
}

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

	fmt.Println("")

	for i := 0; i < len(instances); i++ {
		instance := *instances[i].Instances[0]

		//instanceID := *instance.InstanceId
		//awsInternal.PrintMessage("Preparing the replace of instance-id: " + instanceID)

		fmt.Println(instance)

		/*
			oldInstance := InstanceEc2{
				InstanceType: *instance.InstanceType,
			}*

			fmt.Printf("%+v\n", oldInstance)
			os.Exit(1)  */
	}
}

/*
{
  AmiLaunchIndex: 0,
  Architecture: "x86_64",
  BlockDeviceMappings: [{
      DeviceName: "/dev/sda1",
      Ebs: {
        AttachTime: 2019-01-31 01:54:36 +0000 UTC,
        DeleteOnTermination: true,
        Status: "attached",
        VolumeId: "vol-0a78a7b84a7ef0e0b"
      }
    }],
  ClientToken: "daf5a286-75a5-1a62-d164-a4870c05bb43_subnet-058c1147b1354e8a0_1",
  CpuOptions: {
    CoreCount: 1,
    ThreadsPerCore: 1
  },
  EbsOptimized: false,
  EnaSupport: true,
  HibernationOptions: {
    Configured: false
  },
  Hypervisor: "xen",
  ImageId: "ami-0ca6dfcefdde35d25",
  InstanceId: "i-09ca250b1235916b3",
  InstanceType: "t2.micro",
  LaunchTime: 2019-01-31 01:54:36 +0000 UTC,
  Monitoring: {
    State: "enabled"
  },
  NetworkInterfaces: [{
      Attachment: {
        AttachTime: 2019-01-31 01:54:36 +0000 UTC,
        AttachmentId: "eni-attach-065bbcf634bac76da",
        DeleteOnTermination: true,
        DeviceIndex: 0,
        Status: "attached"
      },
      Description: "",
      Groups: [{
          GroupId: "sg-0d9a04b5f4c4b6b89",
          GroupName: "app-platform"
        }],
      MacAddress: "0e:b5:a4:8e:15:56",
      NetworkInterfaceId: "eni-05f04650f22aca0d1",
      OwnerId: "352631681906",
      PrivateDnsName: "ip-172-31-11-80.ec2.internal",
      PrivateIpAddress: "172.31.11.80",
      PrivateIpAddresses: [{
          Primary: true,
          PrivateDnsName: "ip-172-31-11-80.ec2.internal",
          PrivateIpAddress: "172.31.11.80"
        }],
      SourceDestCheck: true,
      Status: "in-use",
      SubnetId: "subnet-058c1147b1354e8a0",
      VpcId: "vpc-d69537ac"
    }],
  Placement: {
    AvailabilityZone: "us-east-1a",
    GroupName: "",
    Tenancy: "default"
  },
  PrivateDnsName: "ip-172-31-11-80.ec2.internal",
  PrivateIpAddress: "172.31.11.80",
  ProductCodes: [{
      ProductCodeId: "aw0evgkw8e5c1q413zgy5pjce",
      ProductCodeType: "marketplace"
    }],
  PublicDnsName: "",
  RootDeviceName: "/dev/sda1",
  RootDeviceType: "ebs",
  SecurityGroups: [{
      GroupId: "sg-0d9a04b5f4c4b6b89",
      GroupName: "app-platform"
    }],
  SourceDestCheck: true,
  State: {
    Code: 16,
    Name: "running"
  },
  StateTransitionReason: "",
  SubnetId: "subnet-058c1147b1354e8a0",
  Tags: [{
      Key: "aws:autoscaling:groupName",
      Value: "app-platform"
    }],
  VirtualizationType: "hvm",
  VpcId: "vpc-d69537ac"
}*/
