package main

import (
    "fmt"
    "time"
    "os"
    "github.com/crowdmob/goamz/aws"
    "github.com/crowdmob/goamz/cloudwatch"
)

type credentials struct {
    AccessKeyId string
    SecretAccessKey string
    Region string
}

func main() {
    params := &credentials{AccessKeyId: "an access key id", SecretAccessKey: "a secret key", Region: "eu-west-1"}
    region := aws.Regions[params.Region]
    now := time.Now()

    auth, err := aws.GetAuth(params.AccessKeyId, params.SecretAccessKey, "", now)
    if err != nil {
       fmt.Printf("Error: %+v\n", err)
       os.Exit(1)
    }
    cw, err := cloudwatch.NewCloudWatch(auth, region.CloudWatchServicepoint)
    request := &cloudwatch.ListMetricsRequest{Namespace: "AWS/EC2"}

    response, err := cw.ListMetrics(request)
    if err == nil {
        fmt.Printf("%+v\n", response)
    } else {
        fmt.Printf("Error: %+v\n", err)
    }
}
