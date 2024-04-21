package constants

import (
	"fmt"
	"time"
)

const (
	IdentityPoolId     string        = "eu-west-1:bce21571-e3a6-47a4-8032-fd015213405f"
	Region             string        = "eu-west-1"
	PoolID             string        = "eu-west-1_0GLV9KO1p"
	ClientID           string        = "6timr8knllr4frovfvq8r2o6oo"
	WssUrl             string        = "wss://4rpyi2ae3f.execute-api.eu-west-1.amazonaws.com/prodbd"
	Service            string        = "execute-api"
	IdleTimeoutMinutes time.Duration = 10 * time.Minute
	SignWrlFormat                    = "X-Amz-Algorithm=AWS4-HMAC-SHA256&" +
		"X-Amz-Credential=%s" +
		"X-Amz-Date=%s" +
		"X-Amz-Security-Token=%s" +
		"X-Amz-SignedHeaders=host&" +
		"X-Amz-Signature=%s"
)

var CognitoIdp string

func init() {
	CognitoIdp = fmt.Sprintf("cognito-idp.%s.amazonaws.com/%s", Region, PoolID)
}
