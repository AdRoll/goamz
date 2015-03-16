package dynamodb

import (
	"strconv"

	"gopkg.in/check.v1"
)

type BatchSuite struct {
	TableDescriptionT TableDescriptionT
	DynamoDBTest
	WithRange bool
}

func (s *BatchSuite) SetUpSuite(c *check.C) {
	setUpAuth(c)
	s.DynamoDBTest.TableDescriptionT = s.TableDescriptionT
	s.server = New(dynamodb_auth, dynamodb_region)
	pk, err := s.TableDescriptionT.BuildPrimaryKey()
	if err != nil {
		c.Skip(err.Error())
	}
	s.table = s.server.NewTable(s.TableDescriptionT.TableName, pk)

	// Cleanup
	s.TearDownSuite(c)
	_, err = s.server.CreateTable(s.TableDescriptionT)
	if err != nil {
		c.Fatal(err)
	}
	s.WaitUntilStatus(c, "ACTIVE")
}

var batch_suite = &BatchSuite{
	TableDescriptionT: TableDescriptionT{
		TableName: "DynamoDBTestMyTable",
		AttributeDefinitions: []AttributeDefinitionT{
			AttributeDefinitionT{"TestHashKey", "S"},
			AttributeDefinitionT{"TestRangeKey", "N"},
		},
		KeySchema: []KeySchemaT{
			KeySchemaT{"TestHashKey", "HASH"},
			KeySchemaT{"TestRangeKey", "RANGE"},
		},
		ProvisionedThroughput: ProvisionedThroughputT{
			ReadCapacityUnits:  1,
			WriteCapacityUnits: 1,
		},
	},
	WithRange: true,
}

var _ = check.Suite(batch_suite)

func (s *BatchSuite) TestBatchGetDocument(c *check.C) {
	numKeys := 3
	keys := make([]*Key, 0, numKeys)
	ins := make([]map[string]interface{}, 0, numKeys)
	outs := make([]map[string]interface{}, numKeys)
	for i := 0; i < numKeys; i++ {
		k := &Key{HashKey: "NewHashKeyVal" + strconv.Itoa(i)}
		if s.WithRange {
			k.RangeKey = strconv.Itoa(12 + i)
		}

		in := map[string]interface{}{
			"Attr1": "Attr1Val" + strconv.Itoa(i),
			"Attr2": 12 + i,
		}

		if i%2 == 0 { // only add the even keys
			if err := s.table.PutDocument(k, in); err != nil {
				c.Fatal(err)
			}
		}

		keys = append(keys, k)
		ins = append(ins, in)
	}

	err, errs := s.table.BatchGetDocument(keys, true, outs)
	if err != nil {
		c.Fatal(err)
	}

	for i := 0; i < numKeys; i++ {
		if i%2 == 0 {
			c.Assert(errs[i], check.Equals, nil)
			c.Assert(outs[i], check.DeepEquals, ins[i])
		} else {
			c.Assert(errs[i], check.Equals, ErrNotFound)
		}
	}
}

func (s *BatchSuite) TestBatchGetDocumentTyped(c *check.C) {
	type myInnterStruct struct {
		List []interface{}
	}
	type myStruct struct {
		Attr1  string
		Attr2  int64
		Nested myInnterStruct
	}

	numKeys := 3
	keys := make([]*Key, 0, numKeys)
	ins := make([]myStruct, 0, numKeys)
	outs := make([]myStruct, numKeys)

	for i := 0; i < numKeys; i++ {
		k := &Key{HashKey: "NewHashKeyVal" + strconv.Itoa(i)}
		if s.WithRange {
			k.RangeKey = strconv.Itoa(12 + i)
		}

		in := myStruct{
			Attr1:  "Attr1Val" + strconv.Itoa(i),
			Attr2:  1000000 + int64(i),
			Nested: myInnterStruct{[]interface{}{true, false, nil, "some string", 3.14}},
		}

		if i%2 == 0 { // only add the even keys
			if err := s.table.PutDocument(k, in); err != nil {
				c.Fatal(err)
			}
		}

		keys = append(keys, k)
		ins = append(ins, in)
	}

	err, errs := s.table.BatchGetDocument(keys, true, outs)
	if err != nil {
		c.Fatal(err)
	}

	for i := 0; i < numKeys; i++ {
		if i%2 == 0 {
			c.Assert(errs[i], check.Equals, nil)
			c.Assert(outs[i], check.DeepEquals, ins[i])
		} else {
			c.Assert(errs[i], check.Equals, ErrNotFound)
		}
	}
}
