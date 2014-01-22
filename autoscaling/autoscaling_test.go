package autoscaling

import (
	"github.com/crowdmob/goamz/aws"
	"testing"
)

func TestBasicGroupRequest(t *testing.T) {
	awsAuth, err := aws.EnvAuth()
	if err != nil {
		t.Fatalf("AWS environment variables not set : %v\n", err)
	} else {
		as := New(awsAuth, aws.USWest2)
		groupResp, err := as.DescribeAutoScalingGroups(nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(groupResp.AutoScalingGroups) > 0 {
			firstGroup := groupResp.AutoScalingGroups[0]
			if len(firstGroup.AutoScalingGroupName) > 0 {
				t.Logf("Found AutoScaling group %s\n",
					firstGroup.AutoScalingGroupName)
			}
		}
	}
}

func TestAutoScalingGroup(t *testing.T) {
	// Launch configuration test config
	var lc LaunchConfiguration
	lc.LaunchConfigurationName = "LConf1"
	lc.ImageId = "ami-03e47533" // Octave debian ami
	lc.KernelId = "aki-98e26fa8"
	lc.KeyName = "testAWS" // Replace with valid key for your account
	lc.InstanceType = "m1.small"

	// AutoScalingGroup test config
	var asg AutoScalingGroup
	asg.AutoScalingGroupName = "ASGTest1"
	asg.LaunchConfigurationName = lc.LaunchConfigurationName
	asg.DefaultCooldown = 300
	asg.HealthCheckGracePeriod = 300
	asg.DesiredCapacity = 1
	asg.MinSize = 1
	asg.MaxSize = 5
	asg.AvailabilityZones = []string{"us-west-2a"}

	// Parameters for setting desired capacity to 1
	var sp1 SetDesiredCapacityRequestParams
	sp1.AutoScalingGroupName = asg.AutoScalingGroupName
	sp1.DesiredCapacity = 1

	// Parameters for setting desired capacity to 2
	var sp2 SetDesiredCapacityRequestParams
	sp2.AutoScalingGroupName = asg.AutoScalingGroupName
	sp2.DesiredCapacity = 2

	awsAuth, err := aws.EnvAuth()
	if err != nil {
		t.Fatalf("AWS environment variables not set : %v\n", err)
	} else {
		// Create the launch configuration
		as := New(awsAuth, aws.USWest2)
		_, err = as.CreateLaunchConfiguration(lc)
		if err != nil {
			t.Fatal(err)
		}

		// Check that we can get the launch configuration details
		_, err = as.DescribeLaunchConfigurations([]string{lc.LaunchConfigurationName})
		if err != nil {
			t.Fatal(err)
		}

		// Create the AutoScalingGroup
		_, err = as.CreateAutoScalingGroup(asg)
		if err != nil {
			t.Fatal(err)
		}

		// Check that we can get the autoscaling group details
		_, err = as.DescribeAutoScalingGroups(nil)
		if err != nil {
			t.Fatal(err)
		}

		// Suspend the scaling processes for the test AutoScalingGroup
		_, err = as.SuspendProcesses(asg, nil)
		if err != nil {
			t.Fatal(err)
		}

		// Resume scaling processes for the test AutoScalingGroup
		_, err = as.ResumeProcesses(asg, nil)
		if err != nil {
			t.Fatal(err)
		}

		// Change the desired capacity from 1 to 2. This will launch a second instance
		_, err = as.SetDesiredCapacity(sp2)
		if err != nil {
			t.Fatal(err)
		}

		// Change the desired capacity from 2 to 1. This will terminate one of the instances
		_, err = as.SetDesiredCapacity(sp1)
		if err != nil {
			t.Fatal(err)
		}

		// Update the max capacity for the scaling group
		asg.MaxSize = 6
		_, err = as.UpdateAutoScalingGroup(asg)
		if err != nil {
			t.Fatal(err)
		}

		// Add a scheduled action to the group
		var psar PutScheduledActionRequestParams
		psar.AutoScalingGroupName = asg.AutoScalingGroupName
		psar.MaxSize = 4
		psar.ScheduledActionName = "SATest1"
		psar.Recurrence = "30 0 1 1,6,12 *"
		_, err = as.PutScheduledUpdateGroupAction(psar)
		if err != nil {
			t.Fatal(err)
		}

		// List the scheduled actions for the group
		var sar ScheduledActionsRequestParams
		sar.AutoScalingGroupName = asg.AutoScalingGroupName
		_, err = as.DescribeScheduledActions(sar)
		if err != nil {
			t.Fatal(err)
		}

		// Delete the test scheduled action from the group
		var dsar DeleteScheduledActionRequestParams
		dsar.AutoScalingGroupName = asg.AutoScalingGroupName
		dsar.ScheduledActionName = psar.ScheduledActionName
		_, err = as.DeleteScheduledAction(dsar)
		if err != nil {
			t.Fatal(err)
		}
	}
}
