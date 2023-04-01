package services

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

const (
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	Sender = "namkatcedrickjumtock3@gmail.com"
	// Replace recipient@example.com with a "To" address. If your account
	// is still in the sandbox, this address must be verified.
	Recipient = "cedrickjumtock@gmail.com"

	// Specify a configuration set. To use a configuration.
	// set, comment the next line and line 92.
	// ConfigurationSet = "ConfigSet".

	// The subject line for the email.
	Subject = "Amazon SES Test (AWS SDK for Go)"

	// The HTML body for the email.
	HTMLBody = "<h1>Amazon SES Test Email (AWS SDK for Go) From Cliqkets.com</h1><p>This email was sent with " +
		"<a href='https://aws.amazon.com/ses/'>Amazon SES</a> using the " +
		"<a href='https://aws.amazon.com/sdk-for-go/'>AWS SDK for Go</a>.</p>"

	// The email body for recipients with non-HTML email clients.
	TextBody = "This email was sent with Amazon SES using the AWS SDK for Go."

	// The character encoding for the email.
	CharSet = "UTF-8"
)

//nolint:funlen
func SendEmail() (bool, error) {
	// Create a new session in the us-west-2 region.
	// Replace us-west-2 with the AWS Region you're using for Amazon SES.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	},
	)
	if err != nil {
		return false, err
	}
	// create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(HTMLBody),
				},
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
		// Uncomment to use a configuration set.
		// ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)
	// Display error messages if they occur.
	if err != nil {
		//nolint:errorlint
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				// logger.Error().Str(ses.ErrCodeMessageRejected, aerr.Error()).Str("Email could not be sent", aerr.Message())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				// logger.Error().Str(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error()).Str("Email could not be sent", aerr.Message())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				// logger.Error().Str(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error()).Str("Error Occure while sending the Email", aerr.Message())
			default:
				// logger.Error().Str("Error occure while sending the Email", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			// logger.Error().Str("Error occure while sending the Email", aerr.Error())
		}

		return false, err
	}

	// logger.Info().Str("Email Sent to address: ", Recipient)
	// logger.Debug().Str("Email process results", fmt.Sprint(result))
	fmt.Sprint(result)

	return true, err
}
