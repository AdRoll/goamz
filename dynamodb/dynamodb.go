package dynamodb

import simplejson "github.com/bitly/go-simplejson"
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

// Error represents an error in an operation with Dynamodb (following goamz/s3)
type Error struct {
	StatusCode	int    // HTTP status code (200, 403, ...)
	Status		string
	Code		string // Dynamodb error code ("MalformedQueryString", ...)
	Message		string // The human-oriented error message
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

func buildError(r *http.Response, jsonBody []byte) error {

	ddbError := Error{
		StatusCode:	r.StatusCode,
		Status:		r.Status,
	}
	// TODO return error if Unmarshal fails?

	json, err := simplejson.NewJson(jsonBody)
	if err != nil {
		log.Printf("Failed to parse body as JSON")
		return err
	}
	ddbError.Message = json.Get("message").MustString()

	// Of the form: com.amazon.coral.validate#ValidationException
	// We only want the last part
	codeStr := json.Get("__type").MustString()
	hashIndex := strings.Index(codeStr, "#")
	if hashIndex > 0 {
		codeStr = codeStr[hashIndex+1:]
	}
	ddbError.Code = codeStr

	return &ddbError
}

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

	// http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/ErrorHandling.html
	// "A response code of 200 indicates the operation was successful."
	if resp.StatusCode != 200 {
		ddbErr := buildError(resp, body)
		return nil, ddbErr
	}

	return body, nil
}

func target(name string) string {
	return "DynamoDB_20120810." + name
}
