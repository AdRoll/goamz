package dynamodb

import (
  "strconv"
)

const (
	TYPE_STRING                         = "S"
	TYPE_NUMBER                         = "N"
	TYPE_BINARY                         = "B"

  TYPE_STRING_SET                     = "SS"
  TYPE_NUMBER_SET                     = "NS"
  TYPE_BINARY_SET                     = "BS"

  COMPARISON_EQUAL                    = "EQ"
  COMPARISON_NOT_EQUAL                = "NE"
 	COMPARISON_LESS_THAN_OR_EQUAL       = "LE"
 	COMPARISON_LESS_THAN                = "LT"
  COMPARISON_GREATER_THAN_OR_EQUAL    = "GE"
 	COMPARISON_GREATER_THAN             = "GT"
  COMPARISON_ATTRIBUTE_EXISTS         = "NOT_NULL"
  COMPARISON_ATTRIBUTE_DOES_NOT_EXIST = "NULL"
  COMPARISON_CONTAINS                 = "CONTAINS"
  COMPARISON_DOES_NOT_CONTAIN         = "NOT_CONTAINS"
 	COMPARISON_BEGINS_WITH              = "BEGINS_WITH"
  COMPARISON_IN                       = "IN"
 	COMPARISON_BETWEEN                  = "BETWEEN"
)



type PrimaryKey struct {
	KeyAttribute            *Attribute
	RangeAttribute          *Attribute
}

type AttributeDefinitionT struct {
	Name                    string
	Type                    string
}

type KeySchemaT struct {
  AttributeName           string
  KeyType                 string
}

type ProjectionT struct {
  ProjectionType          string
}

type LocalSecondaryIndexT struct {
  IndexName               string
  IndexSizeBytes          int64
  ItemCount               int64
  KeySchema               []KeySchemaT
  Projection              ProjectionT
}

type ProvisionedThroughputT struct {
  NumberOfDecreasesToday  int64
  ReadCapacityUnits       int64
  WriteCapacityUnits      int64
}

type TableDescription struct {
  AttributeDefinitions    []AttributeDefinitionT
  CreationDateTime        float64
  ItemCount               int64
  KeySchema               KeySchemaT
  LocalSecondaryIndexes   []LocalSecondaryIndexT
  ProvisionedThroughput   ProvisionedThroughputT
	TableName               string
  TableSizeBytes          int64
	TableStatus             string
}

type Attribute struct {
	Type                    string
	Name                    string
	Value                   string
}

type AttributeComparison struct {
  AttributeName           string
  ComparisonOperator      string
  AttributeValueList      []Attribute // contains attributes with only types and names (value ignored)
}



func NewEqualInt64AttributeComparison(attributeName string, equalToValue int64) *AttributeComparison {
  numeric := NewNumericAttribute(attributeName, strconv.FormatInt(equalToValue, 10))
  return &AttributeComparison{attributeName,
    COMPARISON_EQUAL,
    []Attribute{ *numeric },
  }
}

func NewEqualStringAttributeComparison(attributeName string, equalToValue string) *AttributeComparison {
  str := NewStringAttribute(attributeName, equalToValue)
  return &AttributeComparison{attributeName,
    COMPARISON_EQUAL,
    []Attribute{ *str },
  }
}

func NewStringAttribute(name string, value string) *Attribute {
	return &Attribute{TYPE_STRING,
		name,
		value,
	}
}

func NewNumericAttribute(name string, value string) *Attribute {
	return &Attribute{TYPE_NUMBER,
		name,
		value,
	}
}

func NewBinaryAttribute(name string, value string) *Attribute {
	return &Attribute{TYPE_BINARY,
		name,
		value,
	}
}

func (k *PrimaryKey) HasRange() bool {
	return k.RangeAttribute != nil
}

// Useful when you may have many goroutines using a primary key, so they don't fuxor up your values.
func (k *PrimaryKey) Clone(h string, r string) []Attribute {
	pk := &Attribute{ k.KeyAttribute.Type,
		k.KeyAttribute.Name,
		h,
	}

	result := []Attribute{*pk}

	if k.HasRange() {
		rk := &Attribute{ k.RangeAttribute.Type,
			k.RangeAttribute.Name,
			r,
		}

		result = append(result, *rk)
	}

	return result
}
