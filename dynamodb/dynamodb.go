package dynamodb

import (
	"fmt"
	"github.com/alimoeeny/goamz/aws"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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

// ALI
// func (s *Server) QueryServer(target string, query *Query) ([]byte, error) {
// 	return s.queryServer(target, query)
// }

func (s *Server) queryServer(target string, query *Query) ([]byte, error) {
	data := strings.NewReader(query.String())

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// 	// 	s := "{
	// 	//     \"TableName\": \"CSUsersEmail\",
	// 	//     \"IndexName\": \"LastPostIndex\",
	// 	//     \"Select\": \"ALL_ATTRIBUTES\",
	// 	//     \"Limit\":3,
	// 	//     \"ConsistentRead\": true,
	// 	//     \"KeyConditions\": {
	// 	//         \"LastPostDateTime\": {
	// 	//             \"AttributeValueList\": [
	// 	//                 {
	// 	//                     \"S\": \"20130101\"
	// 	//                 },
	// 	//                 {
	// 	//                     \"S\": \"20130115\"
	// 	//                 }
	// 	//             ],
	// 	//             \"ComparisonOperator\": \"BETWEEN\"
	// 	//         },
	// 	//         \"ForumName\": {
	// 	//             \"AttributeValueList\": [
	// 	//                 {
	// 	//                     \"S\": \"Amazon DynamoDB\"
	// 	//                 }
	// 	//             ],
	// 	//             \"ComparisonOperator\": \"EQ\"
	// 	//         }
	// 	//     },
	// 	//     \"ReturnConsumedCapacity\": \"TOTAL\"
	// 	// }"

	// 	sdata := `{
	//     "TableName": "CSUsersEmail",
	//     "Select": "ALL_ATTRIBUTES",
	//     "Limit": 3,
	//     "ConsistentRead": true,
	//     "KeyConditions": {
	//         "PK_EMAIL": {
	//             "AttributeValueList": [
	//                 {
	//                     "S": "a"
	//                 },
	//                 {
	//                     "S": "z"
	//                 }
	//             ],
	//             "ComparisonOperator": "BETWEEN"
	//         }
	//     },
	//     "ReturnConsumedCapacity": "TOTAL"
	// }`

	// 	data = strings.NewReader(sdata)

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	hreq, err := http.NewRequest("POST", s.Region.DynamoDBEndpoint+"/", data)
	if err != nil {
		return nil, err
	}

	hreq.Header.Set("Content-Type", "application/x-amz-json-1.0")

	//ALI
	if s.Auth.SecurityToken != "" {
		hreq.Header.Set("X-Amz-Security-Token", s.Auth.SecurityToken)
		//fmt.Printf("Ali: SecToken = %s \n", s.Auth.SecurityToken)
	}

	hreq.Header.Set("X-Amz-Date", time.Now().UTC().Format(aws.ISO8601BasicFormat))
	hreq.Header.Set("X-Amz-Target", target)

	signer := aws.NewV4Signer(s.Auth, "dynamodb", s.Region)
	signer.Sign(hreq)

	resp, err := http.DefaultClient.Do(hreq)

	if err != nil {
		fmt.Printf("Error calling Amazon")
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Could not read response body")
		return nil, err
	}

	return body, nil
}

func target(name string) string {
	return "DynamoDB_20111205." + name
}
