package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/boilingdata/boilingdata/pkg/constants"
)

type AwsCredentials struct {
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	CredentialScope string
}

func (s *Service) GetSignedWssHeader(token string) (http.Header, error) {
	creds, err := getAwsCredentialss(token)
	if err != nil {
		return nil, err
	}
	header, err := getSignedHeaders(creds)
	if err != nil {
		log.Printf("Error getting singned url headers: " + err.Error())
		return nil, err
	}
	return header, err
}

func (s *Service) GetSignedWssUrl(headers http.Header) (string, error) {
	credential, signature, err := extractCredentialAndSignature(headers["Authorization"][0])
	if err != nil {
		log.Printf("Error Extracting Credential and Signature: " + err.Error())
		return "", err
	}
	signedUrl := constants.WssUrl + "?" + fmt.Sprintf(constants.SignWrlFormat, url.QueryEscape(credential)+"&",
		url.QueryEscape(headers["X-Amz-Date"][0])+"&", url.QueryEscape(headers["X-Amz-Security-Token"][0])+"&", url.QueryEscape(signature))
	return signedUrl, nil
}

func extractCredentialAndSignature(header string) (string, string, error) {
	credentialStart := strings.Index(header, "Credential=")
	if credentialStart == -1 {
		return "", "", fmt.Errorf("credential not found in header")
	}
	credentialStart += len("Credential=")
	credentialEnd := strings.Index(header[credentialStart:], ",") + credentialStart
	if credentialEnd == -1 {
		return "", "", fmt.Errorf("credential end not found in header")
	}

	signatureStart := strings.Index(header, "Signature=")
	if signatureStart == -1 {
		return "", "", fmt.Errorf("signature not found in header")
	}
	signatureStart += len("Signature=")
	signatureEnd := len(header)

	return header[credentialStart:credentialEnd], header[signatureStart:signatureEnd], nil
}

func getAwsCredentialss(jwtIdToken string) (AwsCredentials, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(constants.Region))
	if err != nil {
		return AwsCredentials{}, fmt.Errorf("failed to load configuration, %v", err)
	}
	cognitoClient := cognitoidentity.NewFromConfig(cfg)

	out, err := cognitoClient.GetId(context.TODO(), &cognitoidentity.GetIdInput{
		IdentityPoolId: aws.String(constants.IdentityPoolId),
		Logins:         map[string]string{constants.CognitoIdp: jwtIdToken},
	})

	if err != nil {
		log.Printf("Error : " + err.Error())
		return AwsCredentials{}, err
	}

	ctx := context.Background()
	credRes, err := cognitoClient.GetCredentialsForIdentity(ctx, &cognitoidentity.GetCredentialsForIdentityInput{
		IdentityId: out.IdentityId,
		Logins: map[string]string{
			constants.CognitoIdp: jwtIdToken,
		},
	})

	if err != nil {
		log.Printf("Error : " + err.Error())
		return AwsCredentials{}, err
	}

	awsCreds := AwsCredentials{
		AccessKeyId:     *credRes.Credentials.AccessKeyId,
		SecretAccessKey: *credRes.Credentials.SecretKey,
		SessionToken:    *credRes.Credentials.SessionToken,
	}

	return awsCreds, nil
}

func (s *Service) GetAWSSingingHeaders(urlString string) (http.Header, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	// Extract query parameters
	query := u.Query()

	// Extract date
	amzDate := query.Get("X-Amz-Date")

	// Extract signed headers

	// signedHeaders := query.Get("X-Amz-SignedHeaders")

	// Extract algorithm
	algorithm := query.Get("X-Amz-Algorithm")

	// Extract credential
	credential := query.Get("X-Amz-Credential")

	// Extract security token
	securityToken := query.Get("X-Amz-Security-Token")

	// Extract signature
	signature := query.Get("X-Amz-Signature")

	// Prepare headers
	headers := map[string][]string{
		"Authorization": {
			fmt.Sprintf("%s Credential=%s, SignedHeaders=%s, Signature=%s", algorithm, credential, "host;x-amz-date;x-amz-security-token,", signature),
		},
		"X-Amz-Date":           {amzDate},
		"X-Amz-Security-Token": {securityToken},
	}

	return headers, nil
}

func getSignedHeaders(creds AwsCredentials) (http.Header, error) {
	// Create a signer with the given AWS credentials
	signer := v4.NewSigner(credentials.NewStaticCredentials(creds.AccessKeyId, creds.SecretAccessKey, creds.SessionToken))
	wsURL := constants.WssUrl
	req, err := http.NewRequest("GET", wsURL, nil)
	if err != nil {
		return nil, err
	}
	// Sign the request
	_, err = signer.Sign(req, nil, constants.Service, constants.Region, time.Now())
	if err != nil {
		log.Println("Error signing request:", err)
		return nil, nil
	}
	// Return the signed URL
	return req.Header, err
}
