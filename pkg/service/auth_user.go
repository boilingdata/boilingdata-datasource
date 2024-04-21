package service

import (
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/boilingdata/boilingdata/pkg/constants"
	"github.com/boilingdata/boilingdata/pkg/wsclient"
)

type Service struct {
	Wsc      *wsclient.WSSClient
	Response []Response
}

var AuthResult *cognitoidentityprovider.AuthenticationResultType
var timeWhenLastJwtTokenWasRecieved time.Time

func (s *Service) AuthenticateUser(username, password string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(constants.Region)},
	)

	if err != nil {
		return "", err
	}

	cognitoClient := cognitoidentityprovider.New(sess)
	var authInput *cognitoidentityprovider.InitiateAuthInput
	if s.IsUserLoggedIn() && !s.IsTokenExpired() {
		return *AuthResult.IdToken, nil
	} else if s.IsUserLoggedIn() {
		log.Println("Token expired, Getting token with refresh token..")
		authInput = &cognitoidentityprovider.InitiateAuthInput{
			AuthFlow: aws.String("REFRESH_TOKEN_AUTH"),
			AuthParameters: map[string]*string{
				"REFRESH_TOKEN": aws.String(*AuthResult.RefreshToken),
				"POOL_ID":       aws.String(constants.PoolID),
			},
			ClientId: aws.String(constants.ClientID),
		}
	} else {
		log.Println("Logging in..")
		// Authenticate user
		authInput = &cognitoidentityprovider.InitiateAuthInput{
			AuthFlow: aws.String("USER_PASSWORD_AUTH"),
			AuthParameters: map[string]*string{
				"USERNAME": aws.String(username),
				"PASSWORD": aws.String(password),
				"POOL_ID":  aws.String(constants.PoolID),
			},
			ClientId: aws.String(constants.ClientID),
		}
	}
	timeWhenLastJwtTokenWasRecieved = time.Now()
	authOutput, err := cognitoClient.InitiateAuth(authInput)
	if err != nil {
		AuthResult = nil
		return "", err
	}
	AuthResult = authOutput.AuthenticationResult
	// Handle MFA challenges if required
	if authOutput.ChallengeName != nil {
		switch *authOutput.ChallengeName {
		case "SMS_MFA":
			mfaCode, err := promptMFA("Please enter MFA (sms)")
			if err != nil {
				return "", err
			}
			err = sendMFA(cognitoClient, authOutput.Session, mfaCode, "SMS_MFA")
			if err != nil {
				return "", err
			}
		case "SOFTWARE_TOKEN_MFA":
			mfaCode, err := promptMFA("Please enter MFA (totp)")
			if err != nil {
				return "", err
			}
			err = sendMFA(cognitoClient, authOutput.Session, mfaCode, "SOFTWARE_TOKEN_MFA")
			if err != nil {
				return "", err
			}
		}
	}

	// Handle newPasswordRequired if required
	if authOutput.AuthenticationResult == nil {
		newPassword, err := promptPassword("Please enter new password")
		if err != nil {
			return "", err
		}
		// Assume, need to provide new password
		// we might need additional logic to determine if new password is required
		return completeNewPasswordChallenge(cognitoClient, authOutput.Session, newPassword)
	}
	// Authentication successful
	log.Println("Authentication successful")
	return *authOutput.AuthenticationResult.IdToken, nil
}

func (s *Service) IsUserLoggedIn() bool {
	if AuthResult != nil && AuthResult.IdToken != nil {
		return true
	}
	return false
}
func (s *Service) IsTokenExpired() bool {
	if AuthResult != nil && AuthResult.ExpiresIn != nil {
		expirationTime := timeWhenLastJwtTokenWasRecieved.Add(time.Second * time.Duration(*AuthResult.ExpiresIn))
		if !time.Now().After(expirationTime) {
			return false
		}
	}
	return true
}

func promptMFA(promptMsg string) (string, error) {
	// Implement logic for prompting MFA from the user
	return "", errors.New("Prompting for MFA not implemented")
}

func sendMFA(client *cognitoidentityprovider.CognitoIdentityProvider, session *string, mfaCode, mfaType string) error {
	input := &cognitoidentityprovider.RespondToAuthChallengeInput{
		ChallengeName: aws.String("SMS_MFA"), // or "SOFTWARE_TOKEN_MFA"
		ClientId:      aws.String("YOUR_CLIENT_ID"),
		Session:       session,
		ChallengeResponses: map[string]*string{
			"SMS_MFA_CODE": aws.String(mfaCode), // or "SOFTWARE_TOKEN_MFA_CODE"
		},
	}
	_, err := client.RespondToAuthChallenge(input)
	return err
}

func promptPassword(promptMsg string) (string, error) {
	// Implement logic for prompting new password from the user
	return "", errors.New("Prompting for new password not implemented")
}

func completeNewPasswordChallenge(client *cognitoidentityprovider.CognitoIdentityProvider, session *string, newPassword string) (string, error) {
	// Assume that you need to complete new password challenge
	// we might need additional logic here
	return "", nil
}
