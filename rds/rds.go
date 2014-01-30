package rds

import (
	"encoding/xml"
	"github.com/crowdmob/goamz/aws"
	"log"
	"net/http/httputil"
	"strconv"
)

const debug = true

// The RDS type encapsulates operations within a specific EC2 region.
type RDS struct {
	Service aws.AWSService
}

// New creates a new RDS Client.
func New(auth aws.Auth, region aws.ServiceInfo) (*RDS, error) {
	service, err := aws.NewService(auth, region)
	if err != nil {
		return nil, err
	}
	return &RDS{
		Service: service,
	}, nil
}

// ----------------------------------------------------------------------------
// Request dispatching logic.

// query dispatches a request to the RDS API signed with a version 2 signature
func (rds *RDS) query(method, path string, params map[string]string, resp interface{}) error {
	// Add basic RDS param
	params["Version"] = "2010-01-01"

	r, err := rds.Service.Query(method, path, params)
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
		return rds.Service.BuildError(r)
	}
	err = xml.NewDecoder(r.Body).Decode(resp)
	return err
}

// Response to a DescribeDBInstances request
//
// See http://goo.gl/KSPlAl for more details.
type DescribeDBInstancesResp struct {
	DBInstances []DBInstance `xml:"DescribeDBInstancesResult>DBInstances"` // The list of database instances
	Marker      string       `xml:"DescribeDBInstancesResult>Marker"`      // An optional pagination token provided by a previous request
	RequestId   string       `xml:"ResponseMetadata>RequestId"`
}

// DescribeDBInstances - Returns a description of each Database Instance
// Supports pagination by using the "Marker" parameter, and "maxRecords" for subsequent calls
// Unfortunately RDS does not currently support filtering
//
// See http://goo.gl/lzZMyz for more details.
func (rds *RDS) DescribeDBInstances(id string, maxRecords int, marker string) (*DescribeDBInstancesResp, error) {

	params := aws.MakeParams("DescribeDBInstances")

	if id != "" {
		params["DBInstanceIdentifier"] = id
	}

	if maxRecords != 0 {
		params["MaxRecords"] = strconv.Itoa(maxRecords)
	}
	if marker != "" {
		params["Marker"] = marker
	}

	resp := &DescribeDBInstancesResp{}
	err := rds.query("POST", "/", params, resp)
	return resp, err
}
