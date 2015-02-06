package kms

//The rersponse body from KMS, based on which action you take
//http://docs.aws.amazon.com/kms/latest/APIReference/API_Operations.html
type DescribeKeyResp struct {
	KeyMetadata struct {
		AWSAccountId	string
		Arn				string
		CreationDate	float64
		Description		string
		Enabled			bool
		KeyId			string
		KeyUsage		string
	}
}

type EncryptResp struct {
	CiphertextBlob		[]byte
	KeyId				string
}

type DecryptResp struct {
	KeyId		string
	Plaintext	[]byte
}

//For some actions, we just only check if it is success by status code. (200)
//1. EnableKey
//2. DisableKey
