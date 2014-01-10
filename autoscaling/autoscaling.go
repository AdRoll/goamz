package autoscaling

import (
	"encoding/xml"
	"fmt"
	"github.com/JonPulfer/goamz/aws"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
)

const debug = true

var timeNow = time.Now

type AutoScaling struct {
	aws.Auth
	aws.Region
}

type xmlErrors struct {
	RequestId string  `xml:"RequestID"`
	Errors    []Error `xml:"Errors>Error"`
}

type Error struct {
	// HTTP status code (200, 403, ...)
	StatusCode int
	// AutoScaling error code ("UnsupportedOperation", ...)
	Code string
	// The human-oriented error message
	Message   string
	RequestId string `xml:"RequestID"`
}

func (err *Error) Error() string {
	if err.Code == "" {
		return err.Message
	}

	return fmt.Sprintf("%s (%s)", err.Message, err.Code)
}

// Function New creates a new AutoScaling
func New(auth aws.Auth, region aws.Region) *AutoScaling {
	return &AutoScaling{auth, region}
}

func (as *AutoScaling) query(params map[string]string, resp interface{}) error {
	params["Version"] = "2011-01-01"
	params["Timestamp"] = timeNow().In(time.UTC).Format(time.RFC3339)
	endpoint, err := url.Parse(as.Region.AutoScalingEndpoint)
	if err != nil {
		return err
	}
	sign(as.Auth, "GET", endpoint.Path, params, endpoint.Host)
	endpoint.RawQuery = multimap(params).Encode()
	if debug {
		log.Printf("get { %v } -> {\n", endpoint.String())
	}
	r, err := http.Get(endpoint.String())
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if debug {
		dump, _ := httputil.DumpResponse(r, true)
		log.Printf("response:\n")
		log.Printf("%v\n}\n", string(dump))
	}
	if r.StatusCode != 200 {
		return buildError(r)
	}
	err = xml.NewDecoder(r.Body).Decode(resp)
	return err
}

func multimap(p map[string]string) url.Values {
	q := make(url.Values, len(p))
	for k, v := range p {
		q[k] = []string{v}
	}
	return q
}

func makeParams(action string) map[string]string {
	params := make(map[string]string)
	params["Action"] = action
	return params
}

func addParamsList(params map[string]string, label string, ids []string) {
	for i, id := range ids {
		params[label+"."+strconv.Itoa(i+1)] = id
	}
}

func buildError(r *http.Response) error {
	errors := xmlErrors{}
	xml.NewDecoder(r.Body).Decode(&errors)
	var err Error
	if len(errors.Errors) > 0 {
		err = errors.Errors[0]
	}
	err.RequestId = errors.RequestId
	err.StatusCode = r.StatusCode
	if err.Message == "" {
		err.Message = r.Status
	}
	return &err
}

// ----------------------------------------------------------------------------
// Auto Scaling types and related functions.

type AutoScalingGroup struct {
	AutoScalingGroupARN     string     `xml:"AutoScalingGroupARN"`
	AutoScalingGroupName    string     `xml:"AutoScalingGroupName"`
	AvailabilityZones       []string   `xml:"AvailabilityZones>member"`
	CreatedTime             string     `xml:"CreatedTime"`
	DefaultCooldown         int        `xml:"DefaultCooldown"`
	DesiredCapacity         int        `xml:"DesiredCapacity"`
	HealthCheckGracePeriod  int        `xml:"HealthCheckGracePeriod"`
	HealthCheckType         string     `xml:"HealthCheckType"`
	Instances               []Instance `xml:"Instances>member"`
	LaunchConfigurationName string     `xml:"LaunchConfigurationName"`
	LoadBalancerNames       []string   `xml:"LoadBalancerNames>member"`
	MaxSize                 int        `xml:"MaxSize"`
	MinSize                 int        `xml:"MinSize"`
	TerminationPolicies     []string   `xml:"TerminationPolicies>member"`
	VPCZoneIdentifier       string     `xml:"VPCZoneIdentifier"`
}

type Instance struct {
	InstanceId              string `xml:"InstanceId"`
	HealthStatus            string `xml:"HealthStatus"`
	AvailabilityZone        string `xml:"AvailabilityZone"`
	LaunchConfigurationName string `xml:"LaunchConfigurationName"`
	LifecycleState          string `xml:"LifecycleState"`
}

type LaunchConfiguration struct {
	AssociatePublicIpAddress bool     `xml:"AssociatePublicIpAddress"`
	CreatedTime              string   `xml:"CreatedTime"`
	EbsOptimized             bool     `xml:"EbsOptimized"`
	LaunchConfigurationARN   string   `xml:"LaunchConfigurationARN"`
	LaunchConfigurationName  string   `xml:"LaunchConfigurationName"`
	ImageId                  string   `xml:"ImageId"`
	InstanceType             string   `xml:"InstanceType"`
	KernelId                 string   `xml:"KernelId"`
	SecurityGroups           []string `xml:"SecurityGroups>member"`
	KeyName                  string   `xml:"KeyName"`
	UserData                 string   `xml:"UserData"`
}

// Type AutoScalingGroupsResp defines the basic response structure.
type AutoScalingGroupsResp struct {
	RequestId         string             `xml:"ResponseMetadata>RequestId"`
	AutoScalingGroups []AutoScalingGroup `xml:"DescribeAutoScalingGroupsResult>AutoScalingGroups>member"`
}

// Type LaunchConfigurationResp defines the basic response structure for launch configuration
// requests
type LaunchConfigurationResp struct {
	RequestId            string                `xml:"ResponseMetadata>RequestId"`
	LaunchConfigurations []LaunchConfiguration `xml:"DescribeLaunchConfigurationsResult>LaunchConfigurations>member"`
}

// Method AutoScalingGroups returns details about the groups provided in the list. If the list is nil
// information is returned about all the groups in the region.
func (as *AutoScaling) AutoScalingGroups(groupnames []string) (resp *AutoScalingGroupsResp, err error) {
	params := makeParams("DescribeAutoScalingGroups")
	addParamsList(params, "AutoScalingGroupNames", groupnames)
	resp = &AutoScalingGroupsResp{}
	err = as.query(params, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
