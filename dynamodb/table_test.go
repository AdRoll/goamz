package dynamodb_test

import (
	"github.com/crowdmob/goamz/dynamodb"
	"gopkg.in/check.v1"
)

type TableSuite struct {
	TableDescriptionT dynamodb.TableDescriptionT
	DynamoDBTest
}

func (s *TableSuite) SetUpSuite(c *check.C) {
	setUpAuth(c)
	s.DynamoDBTest.TableDescriptionT = s.TableDescriptionT
	s.server = &dynamodb.Server{dynamodb_auth, dynamodb_region}
	pk, err := s.TableDescriptionT.BuildPrimaryKey()
	if err != nil {
		c.Skip(err.Error())
	}
	s.table = s.server.NewTable(s.TableDescriptionT.TableName, pk)

	// Cleanup
	s.TearDownSuite(c)
}

var table_suite = &TableSuite{
	TableDescriptionT: dynamodb.TableDescriptionT{
		TableName: "DynamoDBTestMyTable",
		AttributeDefinitions: []dynamodb.AttributeDefinitionT{
			dynamodb.AttributeDefinitionT{"TestHashKey", "S"},
			dynamodb.AttributeDefinitionT{"TestRangeKey", "N"},
		},
		KeySchema: []dynamodb.KeySchemaT{
			dynamodb.KeySchemaT{"TestHashKey", "HASH"},
			dynamodb.KeySchemaT{"TestRangeKey", "RANGE"},
		},
		ProvisionedThroughput: dynamodb.ProvisionedThroughputT{
			ReadCapacityUnits:  1,
			WriteCapacityUnits: 1,
		},
	},
}

var _ = check.Suite(table_suite)

func (s *TableSuite) TestCreateListTable(c *check.C) {
	status, err := s.server.CreateTable(s.TableDescriptionT)
	if err != nil {
		c.Fatal(err)
	}
	if status != "ACTIVE" && status != "CREATING" {
		c.Error("Expect status to be ACTIVE or CREATING")
	}

	s.WaitUntilStatus(c, "ACTIVE")

	tables, err := s.server.ListTables()
	if err != nil {
		c.Fatal(err)
	}
	c.Check(len(tables), check.Not(check.Equals), 0)
	c.Check(findTableByName(tables, s.TableDescriptionT.TableName), check.Equals, true)
}
