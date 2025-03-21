package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/josepablocastro/poc_remittance"
)

var application *poc_remittance.Application

type RejectPayment struct {
	Number string `json:"number"`
	Reject bool   `json:"reject"`
}

type RDSSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DBHost   string `json:"db_host"`
	DBName   string `json:"db_name"`
	Port     string `json:"port"`
}

func init() {
	region := getEnvironmentValue("AWS_REGION")
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	svc := secretsmanager.NewFromConfig(config)
	data_source_secret := getEnvironmentValue("DATA_SOURCE_SECRET")

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(data_source_secret),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}
	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
		log.Fatal(err.Error())
	}
	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString

	secret := RDSSecret{}
	json.Unmarshal([]byte(secretString), &secret)

	dataSourceURL := fmt.Sprintf("postgresql://%s@%s/%s", url.UserPassword(secret.Username, secret.Password).String(), secret.DBHost, secret.DBName)

	dbAdapter, err := poc_remittance.NewDBAdapter(dataSourceURL)
	application = poc_remittance.NewApplication(dbAdapter)
}

func handler(ctx context.Context, event json.RawMessage) (any, error) {
	request := RejectPayment{}

	err := json.Unmarshal(event, &request)
	if err != nil {
		log.Fatalf("unable to unmarshal request, %v", err)
	}

	log.Printf("REQ %+v", request)

	payment, err := application.AcceptPayment(request.Number, request.Reject)
	log.Printf("RES %+v", payment)

	return payment, nil
}

func main() {
	lambdaFunctionName := getEnvironmentValue("AWS_LAMBDA_FUNCTION_NAME")
	log.Printf("FN %+v", lambdaFunctionName)

	lambda.Start(handler)
}

func getEnvironmentValue(key string) string {
	if os.Getenv(key) == "" {
		log.Fatalf("%s environment variable is missing.", key)
	}
	return os.Getenv(key)
}
