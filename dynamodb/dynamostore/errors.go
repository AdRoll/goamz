package dynamostore

import "fmt"

type TError struct {
	Summary     string
	Description string
}

func (self *TError) Error() string {
	return fmt.Sprintf("DynamoDB Err %d: %s", self.Summary)
}

func MakeError(summary, description string) *TError {
	return &TError{summary, description}
}

var (
	InitGeneralErr       = MakeError("Failed to initialize table", "...")
	DestroyGeneralErr    = MakeError("Failed to destroy table", "...")
	InitUnknownStatusErr = MakeError("Failed to initialize table", "DynamoDB has returned table status that is unknown")
	DeleteErr            = MakeError("Failed to delete record", "...")
	SaveErr              = MakeError("Failed to save record", "...")
	LookupErr            = MakeError("Failed to lookup record", "...")
	NotFoundErr          = MakeError("Record wasnt found", "...")
)
