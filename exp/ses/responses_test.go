package ses

var TestSendEmailError = `
<?xml version="1.0"?>
<ErrorResponse xmlns="http://ses.amazonaws.com/doc/2010-12-01/">
	<Error>
		<Type>Sender</Type>
		<Code>MessageRejected</Code>
		<Message>Email address is not verified.</Message>
	</Error>
	<RequestId>21d1e58d-28b2-4d5f-a974-669c3c67674f</RequestId>
</ErrorResponse>
`

var TestSendEmailOk = `
<?xml version="1.0"?>
<SendEmailResponse xmlns="http://ses.amazonaws.com/doc/2010-12-01/">
	<SendEmailResult>
    	<MessageId>00000131d51d2292-159ad6eb-077c-46e6-ad09-ae7c05925ed4-000000</MessageId>
	</SendEmailResult>
	<ResponseMetadata>
    	<RequestId>d5964849-c866-11e0-9beb-01a62d68c57f</RequestId>
	</ResponseMetadata>
</SendEmailResponse>
`
