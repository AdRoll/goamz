package kms

type DescribeKeyResp struct {
	KeyMetadata struct {
		AWSAccountId	string
		Arn				string
		CreationDate	float64
		Description		string
		Enable			bool
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
