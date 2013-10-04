package dynamodb

import (
	"github.com/crowdmob/goamz/aws"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"log"
)

type Server struct {
	Auth   aws.Auth
	Region aws.Region
}

/*
type Query struct {
	Query string
}
*/

/*
func NewQuery(queryParts []string) *Query {
	return &Query{
		"{" + strings.Join(queryParts, ",") + "}",
	}
}
*/

func (s *Server) queryServer(target string, query *Query) ([]byte, error) {
	data := strings.NewReader(query.String())
	hreq, err := http.NewRequest("POST", s.Region.DynamoDBEndpoint+"/", data)
	if err != nil {
		return nil, err
	}

	hreq.Header.Set("Content-Type", "application/x-amz-json-1.0")
	hreq.Header.Set("X-Amz-Date", time.Now().UTC().Format(aws.ISO8601BasicFormat))
	hreq.Header.Set("X-Amz-Target", target)

	signer := aws.NewV4Signer(s.Auth, "dynamodb", s.Region)
	signer.Sign(hreq)

	resp, err := http.DefaultClient.Do(hreq)

	if err != nil {
		log.Printf("Error calling Amazon")
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Could not read response body")
		return nil, err
	}

	return body, nil
}

func target(name string) string {
	return "DynamoDB_20120810." + name
}
