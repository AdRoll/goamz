package iam_test

// http://goo.gl/EUIvl
var CreateUserExample = `
<CreateUserResponse>
   <CreateUserResult>
      <User>
         <Path>/division_abc/subdivision_xyz/</Path>
         <UserName>Bob</UserName>
         <UserId>AIDACKCEVSQ6C2EXAMPLE</UserId>
         <Arn>arn:aws:iam::123456789012:user/division_abc/subdivision_xyz/Bob</Arn>
     </User>
   </CreateUserResult>
   <ResponseMetadata>
      <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
   </ResponseMetadata>
</CreateUserResponse>
`

var DuplicateUserExample = `
<?xml version="1.0" encoding="UTF-8"?>
<Response>
  <Errors>
    <Error>
	  <Code>EntityAlreadyExists</Code>
      <Message>User already exists.</Message>
    </Error>
  </Errors>
  <RequestID>
    0503f4e9-bbd6-483c-b54f-c4ae9f3b30f4
  </RequestID>
</Response>
`

var CreateAccessKeyExample = `
<CreateAccessKeyResponse>
   <CreateAccessKeyResult>
     <AccessKey>
         <UserName>Bob</UserName>
         <AccessKeyId>AKIAIOSFODNN7EXAMPLE</AccessKeyId>
         <Status>Active</Status>
         <SecretAccessKey>wJalrXUtnFEMI/K7MDENG/bPxRfiCYzEXAMPLEKEY</SecretAccessKey>
      </AccessKey>
   </CreateAccessKeyResult>
   <ResponseMetadata>
      <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
   </ResponseMetadata>
</CreateAccessKeyResponse>
`
