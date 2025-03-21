package:
	rm -f poc_accept_payment_lambda.zip
	rm -f bootstrap
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go && zip poc_accept_payment_lambda.zip bootstrap
