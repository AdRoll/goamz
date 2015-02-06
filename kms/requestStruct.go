package kms

type KMSAction interface {
	ActionName() string
}

type ProduceKeyOpt struct {
	EncryptionContext	map[string]string	`json:",omitempty"`
	GrantTokens			[]string			`json:",omitempty"`
}

//The following structs are the parameters when requesting to AWS KMS
type DescribeKeyInfo struct {
	KeyId	string
}

func (d *DescribeKeyInfo) ActionName() string {
	return "DescribeKey"
}

type EncryptInfo struct {
	KeyId			string
	ProduceKeyOpt
	Plaintext		[]byte
}

func (e *EncryptInfo) ActionName() string {
	return "Encrypt"
}

type DecryptInfo struct {
	CiphertextBlob	[]byte
	ProduceKeyOpt
}

func (d *DecryptInfo) ActionName() string {
	return "Decrypt"
}
