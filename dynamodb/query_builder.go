package dynamodb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	UPDATE_EXPRESSION_ACTION_SET    = "SET"
	UPDATE_EXPRESSION_ACTION_REMOVE = "REMOVE"

	COUNTER_UP   = "UP"
	COUNTER_DOWN = "DOWN"
)

type msi map[string]interface{}

type Query struct {
	buffer msi
}

func NewEmptyQuery() *Query {
	return &Query{msi{}}
}

func NewQuery(t *Table) *Query {
	return NewQueryFor(t.Name)
}

func NewQueryFor(tableName string) *Query {
	q := &Query{msi{"TableName": tableName}}
	return q
}

// This way of specifing the key is used when doing a Get.
// If rangeKey is "", it is assumed to not want to be used
func (q *Query) AddKey(t *Table, key *Key) {
	k := t.Key
	keymap := msi{
		k.KeyAttribute.Name: msi{
			k.KeyAttribute.Type: key.HashKey},
	}
	if k.HasRange() {
		keymap[k.RangeAttribute.Name] = msi{k.RangeAttribute.Type: key.RangeKey}
	}

	q.buffer["Key"] = keymap
}

func (q *Query) AddExclusiveStartKey(t *Table, key *Key) {
	q.buffer["ExclusiveStartKey"] = keyAttributes(t, key)
}

func keyAttributes(t *Table, key *Key) msi {
	k := t.Key

	out := msi{}
	out[k.KeyAttribute.Name] = msi{k.KeyAttribute.Type: key.HashKey}
	if k.HasRange() {
		out[k.RangeAttribute.Name] = msi{k.RangeAttribute.Type: key.RangeKey}
	}
	return out
}

func (q *Query) AddAttributesToGet(attributes []string) {
	if len(attributes) == 0 {
		return
	}

	q.buffer["AttributesToGet"] = attributes
}

func (q *Query) ConsistentRead(c bool) {
	if c == true {
		q.buffer["ConsistentRead"] = "true" //String "true", not bool true
	}
}

func (q *Query) AddGetRequestItems(tableKeys map[*Table][]Key) {
	requestitems := msi{}
	for table, keys := range tableKeys {
		keyslist := []msi{}
		for _, key := range keys {
			keyslist = append(keyslist, keyAttributes(table, &key))
		}
		requestitems[table.Name] = msi{"Keys": keyslist}
	}
	q.buffer["RequestItems"] = requestitems
}

func (q *Query) AddWriteRequestItems(tableItems map[*Table]map[string][][]Attribute) {
	b := q.buffer

	b["RequestItems"] = func() msi {
		out := msi{}
		for table, itemActions := range tableItems {
			out[table.Name] = func() interface{} {
				out2 := []interface{}{}
				for action, items := range itemActions {
					for _, attributes := range items {
						Item_or_Key := map[bool]string{true: "Item", false: "Key"}[action == "Put"]
						out2 = append(out2, msi{action + "Request": msi{Item_or_Key: attributeList(attributes)}})
					}
				}
				return out2
			}()
		}
		return out
	}()
}

func (q *Query) AddCreateRequestTable(description TableDescriptionT) {
	b := q.buffer

	attDefs := []interface{}{}
	for _, attr := range description.AttributeDefinitions {
		attDefs = append(attDefs, msi{
			"AttributeName": attr.Name,
			"AttributeType": attr.Type,
		})
	}
	b["AttributeDefinitions"] = attDefs
	b["KeySchema"] = description.KeySchema
	b["TableName"] = description.TableName
	b["ProvisionedThroughput"] = msi{
		"ReadCapacityUnits":  int(description.ProvisionedThroughput.ReadCapacityUnits),
		"WriteCapacityUnits": int(description.ProvisionedThroughput.WriteCapacityUnits),
	}

	localSecondaryIndexes := []interface{}{}

	for _, ind := range description.LocalSecondaryIndexes {
		localSecondaryIndexes = append(localSecondaryIndexes, msi{
			"IndexName":  ind.IndexName,
			"KeySchema":  ind.KeySchema,
			"Projection": ind.Projection,
		})
	}

	if len(localSecondaryIndexes) > 0 {
		b["LocalSecondaryIndexes"] = localSecondaryIndexes
	}

	globalSecondaryIndexes := []interface{}{}

	for _, ind := range description.GlobalSecondaryIndexes {
		globalSecondaryIndexes = append(globalSecondaryIndexes, msi{
			"IndexName":  ind.IndexName,
			"KeySchema":  ind.KeySchema,
			"Projection": ind.Projection,
			"ProvisionedThroughput": msi{
				"ReadCapacityUnits":  int(ind.ProvisionedThroughput.ReadCapacityUnits),
				"WriteCapacityUnits": int(ind.ProvisionedThroughput.WriteCapacityUnits),
			},
		})
	}

	if len(globalSecondaryIndexes) > 0 {
		b["GlobalSecondaryIndexes"] = globalSecondaryIndexes
	}
}

func (q *Query) AddDeleteRequestTable(description TableDescriptionT) {
	b := q.buffer
	b["TableName"] = description.TableName
}

func (q *Query) AddKeyConditions(comparisons []AttributeComparison) {
	q.buffer["KeyConditions"] = buildComparisons(comparisons)
}

func (q *Query) AddLimit(limit int64) {
	q.buffer["Limit"] = limit
}
func (q *Query) AddSelect(value string) {
	q.buffer["Select"] = value
}

func (q *Query) AddIndex(value string) {
	q.buffer["IndexName"] = value
}

/*
   "ScanFilter":{
       "AttributeName1":{"AttributeValueList":[{"S":"AttributeValue"}],"ComparisonOperator":"EQ"}
   },
*/
func (q *Query) AddScanFilter(comparisons []AttributeComparison) {
	q.buffer["ScanFilter"] = buildComparisons(comparisons)
}

func (q *Query) AddParallelScanConfiguration(segment int, totalSegments int) {
	q.buffer["Segment"] = segment
	q.buffer["TotalSegments"] = totalSegments
}

func buildComparisons(comparisons []AttributeComparison) msi {
	out := msi{}

	for _, c := range comparisons {
		avlist := []interface{}{}
		for _, attributeValue := range c.AttributeValueList {
			avlist = append(avlist, msi{attributeValue.Type: attributeValue.Value})
		}
		out[c.AttributeName] = msi{
			"AttributeValueList": avlist,
			"ComparisonOperator": c.ComparisonOperator,
		}
	}

	return out
}

// The primary key must be included in attributes.
func (q *Query) AddItem(attributes []Attribute) {
	q.buffer["Item"] = attributeList(attributes)
}

func (q *Query) AddUpdates(attributes []Attribute, action string) {
	updates := msi{}
	for _, a := range attributes {
		au := msi{
			"Value": msi{
				a.Type: map[bool]interface{}{true: a.SetValues, false: a.Value}[a.SetType()],
			},
			"Action": action,
		}
		// Delete 'Value' from AttributeUpdates if Type is not Set
		if action == "DELETE" && !a.SetType() {
			delete(au, "Value")
		}
		updates[a.Name] = au
	}

	q.buffer["AttributeUpdates"] = updates
}

func (q *Query) AddExpected(attributes []Attribute) {
	expected := msi{}
	for _, a := range attributes {
		value := msi{}
		if a.Exists != "" {
			value["Exists"] = a.Exists
		}
		// If set Exists to false, we must remove Value
		if value["Exists"] != "false" {
			value["Value"] = msi{a.Type: map[bool]interface{}{true: a.SetValues, false: a.Value}[a.SetType()]}
		}
		expected[a.Name] = value
	}
	q.buffer["Expected"] = expected
}

func attributeList(attributes []Attribute) msi {
	b := msi{}
	for _, a := range attributes {
		//UGH!!  (I miss the query operator)
		b[a.Name] = msi{a.Type: map[bool]interface{}{true: a.SetValues, false: a.Value}[a.SetType()]}
	}
	return b
}

func (q *Query) String() string {
	bytes, _ := json.Marshal(q.buffer)
	return string(bytes)
}

func (a Attribute) ToUpdateExpressionAttribute() UpdateExpressionAttribute {
	ua := UpdateExpressionAttribute{}
	ua.Attribute = a
	return ua
}

// Wrap Attributes and adds flag field to handle counters
type UpdateExpressionAttribute struct {
	Attribute

	// Counter | DESCRIPTION
	// -------------------------------
	// UP      | Increment counter
	// DOWN    | Decrement counter
	// ""      | Do nothing
	Counter string
}

// Wrap Attributes with Operator, used by ConditionExpression
type ConditionExpressionAttribute struct {
	Attribute

	// Operator              | DESCRIPTION
	// -------------------------------
	// =                     | a = b — true if a is equal to b
	// <>                    | true if a is not equal to b
	// <                     | true if a is less than b
	// <=                    | true if a is less than or equal to b
	// >                     | true if a is greater than b
	// >=                    | true if a is greater than or equal to b
	// BETWEEN               | a BETWEEN b AND c - true if a is greater than or equal to b, and less than or equal to c.
	// IN                    | a IN (b, c, d) — true if a is equal to any value in the list — for example, any of b, c or d.
	// attribute_exists      | attribute_exists (path) — true if the attribute at the specified path exists
	// attribute_not_exists  | attribute_not_exists (path) — true if the attribute at the specified path does not exist.
	// begins_with           | begins_with (path, operand) — true if the attribute at the specified path begins with a particular operand.
	// contains              | contains (path, operand) — true if the attribute at the specified path contains a particular operand.
	Operator string
}

func (ca *ConditionExpressionAttribute) build(expressionValuesPrefix string) (condition string, appendExpressionValues []Attribute, err error) {
	singleExpressionValue := func() {
		a := ca.Attribute
		a.Name = expressionValuesPrefix + a.Name
		appendExpressionValues = []Attribute{a}
	}

	multipleExpressionValue := func() {
		appendExpressionValues = []Attribute{}
		for i, av := range ca.SetValues {
			a := (*ca).Attribute
			a.SetValues = nil
			a.Value = av
			a.Name = fmt.Sprintf("%s%d%s", expressionValuesPrefix, i, a.Name)
			appendExpressionValues = append(appendExpressionValues, a)
		}
	}

	switch ca.Operator {
	case "=", "<>", "<", "<=", ">", ">=":
		condition = fmt.Sprintf("%s %s %s%s", ca.Name, ca.Operator, expressionValuesPrefix, ca.Name)
		singleExpressionValue()
	case "attribute_exists", "attribute_not_exists":
		condition = fmt.Sprintf("%s(%s)", ca.Operator, ca.Name)
		singleExpressionValue()
	case "begins_with", "contains":
		condition = fmt.Sprintf("%s(%s, %s%s)", ca.Operator, ca.Name, expressionValuesPrefix, ca.Name)
		singleExpressionValue()
	case "BETWEEN":
		if len(ca.SetValues) == 2 {
			condition = fmt.Sprintf("%s BETWEEN %s0%s AND %s1%s", ca.Name, expressionValuesPrefix, ca.Name, expressionValuesPrefix, ca.Name)
			multipleExpressionValue()
		} else {
			err = fmt.Errorf("BETWEEN operator requires two values of %s attribute", ca.Name)
		}
	case "IN":
		if len(ca.SetValues) > 0 {
			list := make([]string, len(ca.SetValues))
			for i, _ := range ca.SetValues {
				list[i] = fmt.Sprintf("%s%d%s", expressionValuesPrefix, i, ca.Name)
			}

			condition = fmt.Sprintf("%s IN (%s)", ca.Name, concatenate(",", list))
			multipleExpressionValue()
		} else {
			err = fmt.Errorf("IN operator requires at least one value of %s attribute", ca.Name)
		}
	default:
		err = fmt.Errorf("Unknow operator: %s, for attribute: %s", ca.Operator, ca.Name)
	}
	return
}

type ConditionExpression struct {
	Attributes     []*ConditionExpressionAttribute
	SubExpressions []*ConditionExpression

	// Evaluator | DESCRIPTION
	// -------------------------------
	// AND       | a AND b — true if a and b are both true
	// OR        | a OR b — true if either a or b (or both) are true
	// NOT       | NOT a — true if a is false; false if a is true.
	// ""        | empty operator is used with a single Attribute or a single SubExpression
	Evaluator string
}

func (c *ConditionExpression) build(prefix string) (condition string, appendExpressionValues []Attribute, err error) {
	if (c.Attributes != nil && len(c.Attributes) > 0) == (c.SubExpressions != nil && len(c.SubExpressions) > 0) {
		return "", nil, fmt.Errorf("ConditionalExpression should contain exclusively non empty Attributes or SubExpressions list")
	}

	operands := []string{}
	if c.Attributes != nil && len(c.Attributes) > 0 {
		for i, a := range c.Attributes {
			s, ae, err := a.build(fmt.Sprintf("%s%d", prefix, i))
			if err != nil {
				return "", nil, err
			}
			appendExpressionValues = append(appendExpressionValues, ae...)
			operands = append(operands, s)
		}
	} else if c.SubExpressions != nil && len(c.SubExpressions) > 0 {
		for i, a := range c.SubExpressions {
			s, ae, err := a.build(fmt.Sprintf("%d", i))
			if err != nil {
				return "", nil, err
			}
			appendExpressionValues = append(appendExpressionValues, ae...)
			operands = append(operands, fmt.Sprintf("(%s)", s))
		}
	}

	switch c.Evaluator {
	case "AND", "OR":
		condition = concatenate(fmt.Sprintf(" %s ", c.Evaluator), operands)
	case "NOT":
		if len(operands) != 1 {
			return "", nil, fmt.Errorf("NOT evaluator can use a single Attribute or SubExpression")
		}
		condition = fmt.Sprintf("NOT %s", operands[0])
	case "":
		if len(operands) != 1 {
			return "", nil, fmt.Errorf("Empty evaluator is allowed only for single Attribute or SubExpression")
		}
		condition = operands[0]

	default:
		return "", nil, fmt.Errorf("Unknow evaluator :%s", c.Evaluator)
	}

	return condition, appendExpressionValues, nil
}

func concatenate(separator string, list []string) string {
	res := ""
	for _, s := range list {
		if res != "" {
			res = res + separator
		}
		res = res + s
	}
	return res
}

// Append UpdateExpression part of UpdateItem
//
// Action | DESCRIPTION - see http://goo.gl/ufVvpk
// -------------------------------
// SET    | Set one or more attributes
// REMOVE | Remove one or more attributes
func (q *Query) AddUpdateExpression(attributes []UpdateExpressionAttribute, action string) {
	var buffer bytes.Buffer

	switch strings.ToUpper(action) {
	case UPDATE_EXPRESSION_ACTION_SET:
		aCopy := make([]Attribute, len(attributes))
		for i, ac := range attributes {
			aCopy[i] = ac.Attribute
			aCopy[i].Name = ":v" + ac.Name
		}
		q.appendExpressionAttributeValues(aCopy)

		for _, a := range attributes {
			if buffer.Len() > 0 {
				buffer.WriteString(", ")
			}

			if a.Counter == COUNTER_UP && a.Type == TYPE_NUMBER {
				buffer.WriteString(fmt.Sprintf("%v = %v + :v%v", a.Name, a.Name, a.Name))
			} else if a.Counter == COUNTER_DOWN && a.Type == TYPE_NUMBER {
				buffer.WriteString(fmt.Sprintf("%v = %v - :v%v", a.Name, a.Name, a.Name))
			} else {
				if a.Exists == "false" {
					buffer.WriteString(fmt.Sprintf("%v = if_not_exists(%s, :v%v)", a.Name, a.Name, a.Name))
				} else {
					buffer.WriteString(fmt.Sprintf("%v = :v%v", a.Name, a.Name))
				}
			}
		}
	case UPDATE_EXPRESSION_ACTION_REMOVE:
		for _, a := range attributes {
			if buffer.Len() > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteString(a.Name)
		}
	}
	q.buffer["UpdateExpression"] = fmt.Sprintf("%s %s", action, buffer.String())
}

// Append ConditionExpression to query - see: http://goo.gl/UHDOqu
func (q *Query) AddConditionExpression(ca *ConditionExpression) error {
	if condition, ea, err := ca.build(":ca"); err != nil {
		return err
	} else {
		q.appendExpressionAttributeValues(ea)
		q.buffer["ConditionExpression"] = condition
	}
	return nil
}

func (q *Query) appendExpressionAttributeValues(attributes []Attribute) {
	tmp := attributeList(attributes)
	eaMsi := msi{}

	if m, ok := q.buffer["ExpressionAttributeValues"]; ok {
		eaMsi = m.(msi)
	}

	for k, v := range tmp {
		eaMsi[k] = v
	}
	q.buffer["ExpressionAttributeValues"] = eaMsi
}
