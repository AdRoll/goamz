/***** BEGIN LICENSE BLOCK *****
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this file,
# You can obtain one at http://mozilla.org/MPL/2.0/.
#
# The Initial Developer of the Original Code is the Mozilla Foundation.
# Portions created by the Initial Developer are Copyright (C) 2012
# the Initial Developer. All Rights Reserved.
#
# Contributor(s):
#   Ben Bangert (bbangert@mozilla.com)
#
# ***** END LICENSE BLOCK *****/

package cloudwatch

import (
	"errors"
	"github.com/crowdmob/goamz/aws"
	"github.com/feyeleanor/sets"
	"time"
)

// The CloudWatch type encapsulates all the CloudWatch operations in a region.
type CloudWatch struct {
	aws.Auth
	aws.Region
	namespace string
}

type Dimension struct {
	Name  string
	Value string
}

type StatisticSet struct {
	Maximum     float64
	Minimum     float64
	SampleCount float64
	Sum         float64
}

type MetricDatum struct {
	Dimensions      []Dimension
	MetricName      string
	StatisticValues StatisticSet
	Timestamp       time.Time
	Unit            string
	Value           float64
}

type Datapoint struct {
	Average     float64
	Maximum     float64
	Minimum     float64
	SampleCount float64
	Sum         float64
	Timestamp   time.Time
	Unit        string
}

type Params map[string]string

type RequestParams interface {
	// All the params to include in the request
	Params() Params
	// All the params used to generate the request signature
	SignedParams() Params
}

// ResponseMetadata
type ResponseMetadata struct {
	RequestId string  // A unique ID for tracking the request
	BoxUsage  float64 // The measure of machine utilization for this request.
}

type AWSResponse struct {
	ResponseMetadata ResponseMetadata
}

type GetMetricStatisticsRequest struct {
	Dimensions []Dimension
	EndTime    time.Time
	StartTime  time.Time
	MetricName string
	Unit       string
	Period     int
	Statistics []string
}

func (*GetMetricStatisticsRequest) Params() map[string]string {
	p := make(map[string]string)
	return p
}

type GetMetricStatisticsResult struct {
	Datapoints []Datapoint
	Label      string
}

var attempts = aws.AttemptStrategy{
	Min:   5,
	Total: 5 * time.Second,
	Delay: 200 * time.Millisecond,
}

var validUnits = sets.SSet(
	"Seconds",
	"Microseconds",
	"Milliseconds",
	"Bytes",
	"Kilobytes",
	"Megabytes",
	"Gigabytes",
	"Terabytes",
	"Bits",
	"Kilobits",
	"Megabits",
	"Gigabits",
	"Terabits",
	"Percent",
	"Count",
	"Bytes/Second",
	"Kilobytes/Second",
	"Megabytes/Second",
	"Gigabytes/Second",
	"Terabytes/Second",
	"Bits/Second",
	"Kilobits/Second",
	"Megabits/Second",
	"Gigabits/Second",
	"Terabits/Second",
	"Count/Second",
)

var validMetricStatistics = sets.SSet(
	"Average",
	"Sum",
	"SampleCount",
	"Maximum",
	"Minimum",
)

// Create a new CloudWatch object for a given namespace
func New(auth aws.Auth, region aws.Region, namespace string) *CloudWatch {
	return &CloudWatch{auth, region, namespace}
}

// Get statistics for specified metric
//
// If the arguments are invalid or the server returns an error, the error will
// be set and the other values undefined.
func (c *CloudWatch) GetMetricStatistics(req *GetMetricStatisticsRequest) (result *GetMetricStatisticsResult, err error) {
	statisticsSet := sets.SSet(req.Statistics...)
	// Kick out argument errors
	switch {
	case req.EndTime.IsZero():
		err = errors.New("No endTime specified")
	case req.StartTime.IsZero():
		err = errors.New("No startTime specified")
	case req.MetricName == "":
		err = errors.New("No metricName specified")
	case req.Period < 60 || req.Period%60 != 0:
		err = errors.New("Period not 60 seconds or a multiple of 60 seconds")
	case len(req.Statistics) < 1:
		err = errors.New("No statistics supplied")
	case validMetricStatistics.Union(statisticsSet).Len() != validMetricStatistics.Len():
		err = errors.New("Invalid statistic values supplied")
	case req.Unit != "" && !validUnits.Member(req.Unit):
		err = errors.New("Unit is not a valid value")
	}
	if err != nil {
		return
	}

	return
}

func (c *CloudWatch) PutMetricData(metrics []MetricDatum) {

}
