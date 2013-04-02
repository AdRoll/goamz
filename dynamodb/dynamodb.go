//
// goamz:dynamodb - Go packages to interact with the Amazon Web Services' DynamoDB.
//
//   https://wiki.ubuntu.com/goamz
//   https://github.com/crowdmob/goamz
//
// Copyright (c) 2013 CrowdMob Inc.
//
// Written by Matthew Moore <matt@crowdmob.com>
//

package dynamodb

import (
  "io"
	"io/ioutil"
  "log"
  "encoding/json"
  "net"
	"net/http"
	"net/http/httputil"
	"net/url"
  "bytes"
  "strings"
  "strconv"
  "fmt"
  "time"
  

	"launchpad.net/goamz/aws"
)

const debug = true

// The DynamoDB type encapsulates operations with an AWS region.
type DynamoDB struct {
	aws.Auth
	aws.Region
	private byte // Reserve the right of using private data.
}

// The Table type encapsulates operations with an DynamoDB table.
type Table struct {
	*DynamoDB
	TableName string
  KeySchema KeySchemaDescriptor
  ProvisionedThroughput Throughput 
}

type AttributeNameAndType struct {
  AttributeName string
  AttributeType string
}

type Throughput struct {
  ReadCapacityUnits int
  WriteCapacityUnits int
}

type KeySchemaDescriptor struct {
	HashKeyElement AttributeNameAndType
  RangeKeyElement AttributeNameAndType
}

// New creates a new DynamoDB.
func New(auth aws.Auth, region aws.Region) *DynamoDB {
	return &DynamoDB{auth, region, 0}
}


// CreateTable creates a new table.
//
// See http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/API_CreateTable.html for details.
func (t *Table) CreateTable() error {
	headers := map[string][]string{
		"x-amz-target": { "DynamoDB_20111205.CreateTable" },
    "content-type": { "application/x-amz-json-1" },
  }
	
	payld, err := json.Marshal(t)
	if err != nil {
		return err
	}
  
  req := &request{
		method:  "PUT",
		path:    "/",
		headers: headers,
    payload: bytes.NewReader(payld),
	}
  
	return t.DynamoDB.query(req, nil)
}




// ListTables returns an array of all the tables associated with the current account and endpoint. Each Amazon DynamoDB endpoint is entirely independent. For example, if you have two tables called "MyTable," one in dynamodb.us-east-1.amazonaws.com and one in dynamodb.us-west-1.amazonaws.com, they are completely independent and do not share any data. 
//
// See http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/API_ListTables.html

// The ListTablesResponse type holds the results of a List bucket operation.
type ListTablesResponse struct {
	TableNames              []string
	LastEvaluatedTableName  string
}

func (d *DynamoDB) ListTables(exclusiveStartTableName string, limit int) (result *ListTablesResponse, err error) {
  payld := "{\"ExclusiveStartTableName\":\"" + exclusiveStartTableName + "\",\"Limit\":" + strconv.Itoa(limit) + "}"
  
	headers := map[string][]string{
		"x-amz-target": { "DynamoDB_20111205.ListTables" },
    "content-type": { "application/x-amz-json-1" },
  }
  
  
  req := &request{
		method:  "POST",
		path:    "/",
		headers: headers,
    payload: strings.NewReader(payld),
	}
  
  err = d.query(req, result)
	return result, err
}







// -----------------------------------------------------
//   Generic Request Helpers
// -----------------------------------------------------

type request struct {
	method   string
	table    string
	path     string
	signpath string
	params   url.Values
	headers  http.Header
	baseurl  string
	payload  io.Reader
	prepared bool
}


func (req *request) url() (*url.URL, error) {
	u, err := url.Parse(req.baseurl)
	if err != nil {
		return nil, fmt.Errorf("bad dynamodb endpoint URL %q: %v", req.baseurl, err)
	}
	u.RawQuery = req.params.Encode()
	u.Path = req.path
	return u, nil
}

// query prepares and runs the req request.
// If resp is not nil, the JSON data contained in the response
// body will be unmarshalled on it.
func (dynamoDb *DynamoDB) query(req *request, resp interface{}) error {
	err := dynamoDb.prepare(req)
	if err == nil {
		_, err = dynamoDb.run(req, resp)
	}
	return err
}

// prepare sets up req to be delivered to DynamoDB.
func (dynamoDb *DynamoDB) prepare(req *request) error {
	if !req.prepared {
		req.prepared = true
		if req.method == "" {
			req.method = "GET"
		}
		// Copy so they can be mutated without affecting on retries.
		params := make(url.Values)
		headers := make(http.Header)
		for k, v := range req.params {
			params[k] = v
		}
		for k, v := range req.headers {
			headers[k] = v
		}
		req.params = params
		req.headers = headers
		if !strings.HasPrefix(req.path, "/") {
			req.path = "/" + req.path
		}
		req.signpath = req.path
		if req.table != "" {
			req.baseurl = strings.Replace(dynamoDb.Region.EC2Endpoint, "ec2", "dynamodb", -1)
			if req.baseurl == "" {
				// Use the path method to address the table.
				req.baseurl = strings.Replace(dynamoDb.Region.EC2Endpoint, "ec2", "dynamodb", -1)
				req.path = "/" + req.table + req.path
			} else {
				// Just in case, prevent injection.
				if strings.IndexAny(req.table, "/:@") >= 0 {
					return fmt.Errorf("bad DynamoDB table: %q", req.table)
				}
				req.baseurl = strings.Replace(req.baseurl, "${table}", req.table, -1)
			}
			req.signpath = "/" + req.table + req.signpath
		} else  {
			req.baseurl = strings.Replace(dynamoDb.Region.EC2Endpoint, "ec2", "dynamodb", -1)
    }
	}

	// Always sign again as it's not clear how far the
	// server has handled a previous attempt.
	u, err := url.Parse(req.baseurl)
	if err != nil {
		return fmt.Errorf("bad DynamoDB endpoint URL %q: %v", req.baseurl, err)
	}
	req.headers["Host"] = []string{u.Host}
	req.headers["Date"] = []string{time.Now().In(time.UTC).Format(time.RFC1123)}
  
  err = SignV4("dynamodb", dynamoDb.Region.Name, &dynamoDb.Auth, req.method, req.signpath, req.params, req.headers, req.payload)
	if err != nil {
		return fmt.Errorf("Couldn't sign request %v: %v", req, err)
	}
  
	return nil
}

// run sends req and returns the http response from the server.
// If resp is not nil, the XML data contained in the response
// body will be unmarshalled on it.
func (dynamoDb *DynamoDB) run(req *request, resp interface{}) (*http.Response, error) {
	if debug {
		log.Printf("Running DynamoDB request: %#v", req)
	}

	u, err := req.url()
	if err != nil {
		return nil, err
	}

	hreq := http.Request{
		URL:        u,
		Method:     req.method,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Close:      true,
		Header:     req.headers,
	}

	if v, ok := req.headers["Content-Length"]; ok {
		hreq.ContentLength, _ = strconv.ParseInt(v[0], 10, 64)
		delete(req.headers, "Content-Length")
	}
	if req.payload != nil {
  	if debug {
  		log.Printf("PAYLOAD WAS NOT NULL! %#v", req.payload)
  	}

    if hreq.ContentLength == 0 {
      hreq.ContentLength = -1
    }
    // FIXME getting an odd error here with a prepended 2c and trailing 0
    hreq.Body = ioutil.NopCloser(req.payload)
	}

  if debug {
    dump1, _ := httputil.DumpRequestOut(&hreq, true)
		log.Printf("} -> %s\n", dump1)
    
  }

	hresp, err := http.DefaultClient.Do(&hreq)
	if err != nil {
    log.Printf("ERROR!!!!!!!!!!!!!!!!!!!!!!!!!!!! %#v", err)
		return nil, err
	}
	if debug {
		dump2, _ := httputil.DumpResponse(hresp, true)
		log.Printf("} <- %s\n", dump2)
	}
	if hresp.StatusCode != 200 && hresp.StatusCode != 204 {
		return nil, buildError(hresp)
	}
	if resp != nil {
  	json.NewDecoder(hresp.Body).Decode(&resp)
		hresp.Body.Close()
	}
	return hresp, err
}



// Error represents an error in an operation with DynamoDB.
type Error struct {
	StatusCode int    // HTTP status code (200, 403, ...)
	Code       string // EC2 error code ("UnsupportedOperation", ...)
	Message    string // The human-oriented error message
	TableName string
	RequestId  string
	HostId     string
}

func (e *Error) Error() string {
	return e.Message
}

func buildError(r *http.Response) error {
	if debug {
		log.Printf("got error (status code %v)", r.StatusCode)
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("\tread error: %v", err)
		} else {
			log.Printf("\tdata:\n%s\n\n", data)
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	}

	err := Error{}
	// TODO return error if Unmarshal fails?
	json.NewDecoder(r.Body).Decode(&err)
	r.Body.Close()
	err.StatusCode = r.StatusCode
	if err.Message == "" {
		err.Message = r.Status
	}
	if debug {
		log.Printf("err: %#v\n", err)
	}
	return &err
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	switch err {
	case io.ErrUnexpectedEOF, io.EOF:
		return true
	}
	switch e := err.(type) {
	case *net.DNSError:
		return true
	case *net.OpError:
		switch e.Op {
		case "read", "write":
			return true
		}
	case *Error:
		switch e.Code {
		case "InternalError", "NoSuchUpload", "NoSuchBucket":
			return true
		}
	}
	return false
}

func hasCode(err error, code string) bool {
	dynamoerr, ok := err.(*Error)
	return ok && dynamoerr.Code == code
}


/*
 TODO The supported functions in the dynamodb operations list at
 http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/operationlist.html 
 are as follows. It would be great to support all of these at some point as well.

 Implemented functions have a √ before them.

  BatchGetItem
  BatchWriteItem
√ CreateTable
  DeleteItem
  DeleteTable
  DescribeTable
  GetItem
√ ListTables
  PutItem
  Query
  Scan
  UpdateItem
  UpdateTable

*/
