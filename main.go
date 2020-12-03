package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ses"
	"log"
	"strings"
)

const (
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	Sender = "notification@prod.martinyuan.me"

	// Replace recipient@example.com with a "To" address. If your account
	// is still in the sandbox, this address must be verified.
	//Recipient = "chaoyiyuan1@gmail.com"

	// Specify a configuration set. To use a configuration
	// set, comment the next line and line 92.
	//ConfigurationSet = "ConfigSet"

	// The subject line for the email.
	Subject = "Web App Notification"

	// The HTML body for the email.
	/*HtmlBody =  "<h1>Amazon SES Test Email (AWS SDK for Go)</h1><p>This email was sent with " +
		"<a href='https://aws.amazon.com/ses/'>Amazon SES</a> using the " +
		"<a href='prod.martinyuan.me/v1/question/75a29301-ff02-4cab-a14c-8139cfec39c7'>AWS SDK for Go</a>.</p>"*/

	//The email body for recipients with non-HTML email clients.
	//TextBody = "This email was sent with Amazon SES using the AWS SDK for Go."

	// The character encoding for the email.
	CharSet = "UTF-8"
)

func handler(ctx context.Context, snsEvent events.SNSEvent) {
	for _, record := range snsEvent.Records {
		snsRecord := record.SNS
		fmt.Printf("[%s %s] Message = %s \n", record.EventSource, snsRecord.Timestamp, snsRecord.Message)
		sendSNSEmail(snsRecord.Message)
		//tmp := snsRecord.Message
	}
}

func sendSNSEmail(s string) {
	// Create a new session in the us-west-2 region.
	// Replace us-west-2 with the AWS Region you're using for Amazon SES.
	sess, err := session.NewSession(&aws.Config{
		Region:aws.String("us-east-1")},
	)

	// Create an SES session.
	svc := ses.New(sess)

	//arr := strings.Split(s, ",")

	arr := strings.Split(s, ",")
	/*flag := -1
	if arr[0] == "Create an answer" {
		flag = 0
	} else if arr[0] == "Update an answer" {
		flag = 1
	} else if arr[0] == "Delete an answer" {
		flag = 2
	} else {
		flag = -1
	}

	questionId := arr[1]
	qUserEmail := arr[2]
	answerId := arr[3]
	answerTxt := arr[4]

	if flag == -1 {
		fmt.Printf("error: something wrong in message")
		return
	}

	s1 := ""
	if flag == 0 {
		s1 = "Someone created an answer for your question"
	} else if flag == 1 {
		s1 = "Someone updated the answer of your question"
	} else {
		s1 = "Someone deleted the answer of your question"
	}

	link := "prod.martinyuan.me/v1/question/" + questionId
	/*HtmlBody := "<h1>Web App Notification Email </h1><p>The link is " +
		"<a href='" + link + "'>Amazon SES</a> \n" + s1 + "\n" +
		"question_id: " + questionId + "\n" +
		"question_owner_email: " + qUserEmail + "\n" +
		"answer_id: " + answerId + "\n" +
		"answer_text: " + answerTxt + "</p>"*/
	//Recipient := qUserEmail
	/*TextBody := "Web App Notification Email ... \n The link is " +
		link + "." + s1 + "\n" +
		"question_id: " + questionId + "\n" +
		"question_owner_email: " + qUserEmail + "\n" +
		"answer_id: " + answerId + "\n" +
		"answer_text: " + answerTxt*/
	/*TextBody := "arr[0]: " + arr[0] + "\n" +
				"arr[1]: " + arr[1] + "\n" +
				"arr[2]: " + arr[2] + "\n" +
				"arr[3]: " + arr[3] + "\n" +
				"arr[4]: " + arr[4] + "\n"*/
	Recipient := arr[2]
	TextBody := "The user " + arr[0] + ". Please go to this link.\n" +
				"prod.martinyuan.me/v1/question/" + arr[1] + " \n" +
				"question id: " + arr[1] + "\n" +
				"question owner email: " + arr[2] + "\n" +
				"answer id: " + arr[3] + "\n" +
				"answer text: " + arr[4] + "\n"


	//search for email, if already sent, return, otherwise, put in DynamoDB table, and send email
	isExist := searchItemInDynamoDB(TextBody)
	if isExist {
		log.Println("The email has already been sent")
		return
	}

	if err := addItemToDynamoDB(TextBody); err != nil {
		log.Printf("Failed to put email item into DynamoDB table: %v", err)
		return
	}

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{
			},
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				/*Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(HtmlBody),
				},*/
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return
	}

	fmt.Println("Email Sent to address: " + Recipient)
	fmt.Println(result)
}

var svc_db *dynamodb.DynamoDB

func initDBClient() *dynamodb.DynamoDB {
	if svc_db == nil {
		sess, _ := session.NewSession(&aws.Config{
			Region:aws.String("us-east-1")},
		)
		// Create S3 service client
		svc_db = dynamodb.New(sess)
	}

	return svc_db
}

func searchItemInDynamoDB(TextBody string) bool {
	//initialize dynamodb client
	svc_db := initDBClient()

	tableName := "csye6225"

	result, err := svc_db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(TextBody),
			},
		},
	})
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if result.Item == nil {
		log.Println("Search email in dynamodb: false")
		return false
	}

	log.Printf("Got item output: %v", result)
	return true
}

func addItemToDynamoDB(TextBody string) error {
	//initialize dynamodb client
	svc_db := initDBClient()

	tableName := "csye6225"

	item := map[string]*dynamodb.AttributeValue{
		"id": {
			S: aws.String(TextBody),
		},
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	}

	_, err := svc_db.PutItem(input)
	if err != nil {
		log.Printf("Got error calling PutItem: %v\n", err)
		return err
	}

	log.Println("Successfully added email: '" + TextBody + "'")

	return nil
}

func main() {
	lambda.Start(handler)
}