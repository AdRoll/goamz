package main

import (
    "fmt"
    "time"
    "os"
    "github.com/crowdmob/goamz/aws"
    "github.com/crowdmob/goamz/cloudwatch"
)

func main() {
    region := aws.Regions["us-east-1"]  // Any region here
    now := time.Now()

    auth, err := aws.GetAuth("an AccessKeyId", "a SecretAccessKey", "", now)
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
