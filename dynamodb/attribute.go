package dynamodb

const (
	TYPE_STRING = "S"
	TYPE_NUMBER = "N"
	TYPE_BIN    = "B"
)

type PrimaryKey struct {
	KeyAttribute   *Attribute
	RangeAttribute *Attribute
}

type Attribute struct {
	Type  string
	Name  string
	Value string
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
	return &Attribute{TYPE_BIN,
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
