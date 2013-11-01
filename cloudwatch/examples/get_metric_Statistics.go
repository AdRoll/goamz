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
    params := &credentials{AccessKeyId: "your access key id", SecretAccessKey: "your secret key", Region: "eu-west-1"}
    region := aws.Regions[params.Region]
    namespace:= "AWS/ELB"
    dimension  := &cloudwatch.Dimension{Name: "LoadBalancerName", Value: "your_value" }
    metricName := "RequestCount"
    timeRange := 600 // in seconds
    now := time.Now()
    prev := now.Add(time.Duration(timeRange)*time.Second*-1) // 10 minutes

    auth, err := aws.GetAuth(params.AccessKeyId, params.SecretAccessKey, "", now)
    if err != nil {
       fmt.Printf("Error: %+v\n", err)
       os.Exit(1)
    }
    
    cw, err := cloudwatch.NewCloudWatch(auth, region.CloudWatchServicepoint)
    request := &cloudwatch.GetMetricStatisticsRequest {
                                                        Dimensions: []cloudwatch.Dimension{*dimension},
                                                        EndTime: now,
                                                        StartTime: prev,
                                                        MetricName: metricName,
                                                        Unit: "Count", // Not mandatory
                                                        Period: 60,
                                                        Statistics: []string{"Sum"},
                                                        Namespace: namespace,
                                                      }

    response, err := cw.GetMetricStatistics(request)
    if err == nil {
        fmt.Printf("%+v\n", response)
    } else {
        fmt.Printf("Error: %+v\n", err)
    }


}
