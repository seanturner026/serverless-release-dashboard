package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type resetPasswordEvent struct {
	EmailAddress string `json:"email_address"`
	NewPassword  string `json:"new_password"`
	SessionID    string `json:"session_id"`
}

func (app application) resetPassword(e resetPasswordEvent, secretHash string) (string, error) {
	input := &cidp.AdminRespondToAuthChallengeInput{
		ChallengeName: aws.String("NEW_PASSWORD_REQUIRED"),
		ChallengeResponses: map[string]*string{
			"USERNAME":     aws.String(e.EmailAddress),
			"NEW_PASSWORD": aws.String(e.NewPassword),
			"SECRET_HASH":  aws.String(secretHash),
		},
		ClientId:   aws.String(app.config.ClientPoolID),
		UserPoolId: aws.String(app.config.UserPoolID),
		Session:    aws.String(e.SessionID),
	}

	resp, err := app.config.idp.AdminRespondToAuthChallenge(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return "", err
	}

	return *resp.AuthenticationResult.AccessToken, nil
}

func (app application) authResetPasswordHandler(event events.APIGatewayV2HTTPRequest, headers map[string]string) (string, int, map[string]string) {
	e := resetPasswordEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	secretHash := util.GenerateSecretHash(app.config.ClientPoolSecret, e.EmailAddress, app.config.ClientPoolID)
	AccessToken, err := app.resetPassword(e, secretHash)
	if err != nil {
		message := fmt.Sprintf("Error changing user %v password", e.EmailAddress)
		statusCode := 400
		return message, statusCode, headers
	}

	headers["Authorization"] = fmt.Sprintf("Bearer %v", AccessToken)
	message := fmt.Sprintf("User %v changed password successfully", e.EmailAddress)
	statusCode := 200
	return message, statusCode, headers
}
