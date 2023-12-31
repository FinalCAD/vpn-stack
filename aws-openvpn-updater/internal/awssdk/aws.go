package awssdk

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/FinalCAD/vpn-stack/aws-openvpn-updater/internal/settings"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/rs/zerolog/log"
)

const sessionDuration time.Duration = time.Duration(6) * time.Hour
const mailTXT string = `
VPN access :

In this email you'll find a link to download your configuration file for your personal vpn acces.
Do not share this information with anyone, including colleagues.
Those credentials are unique to you and must not be disclosed under any circumstances.
This link will be invalidated during the next hour

`

type AwsSdkConfig struct {
	AwsConfig *settings.Aws
	SdkConfig aws.Config
}

type User struct {
	Name    string
	Account string
}

func (u User) String() string {
	return fmt.Sprintf("[ Name: %v, Account: %v ]", u.Name, u.Account)
}

func CreateIAMConfig(awsconfig *settings.Aws) (*AwsSdkConfig, error) {
	var cfg aws.Config
	var err error

	if awsconfig.Profile != "" {
		log.Debug().Msgf("Aws config using profile: %v", awsconfig.Profile)
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithSharedConfigProfile(awsconfig.Profile), config.WithRegion(awsconfig.Region))
	} else {
		log.Debug().Msgf("Aws config using default role: %v", awsconfig.Profile)
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsconfig.Region))
	}

	if err != nil {
		return nil, err
	}

	if awsconfig.RoleToAssume != "" {
		log.Debug().Msgf("Aws config using assume role: %v", awsconfig.RoleToAssume)
		stsClient := sts.NewFromConfig(cfg)
		stsCreds := stscreds.NewAssumeRoleProvider(stsClient, awsconfig.RoleToAssume, func(o *stscreds.AssumeRoleOptions) {
			o.Duration = sessionDuration
		})
		cfg.Credentials = aws.NewCredentialsCache(stsCreds)
	}

	return &AwsSdkConfig{AwsConfig: awsconfig, SdkConfig: cfg}, nil
}

func (awsSdkCfg *AwsSdkConfig) GetIAMUser() ([]User, error) {
	var users []User
	group := awsSdkCfg.AwsConfig.VpnGroup

	svc := iam.NewFromConfig(awsSdkCfg.SdkConfig)
	resp, err := svc.GetGroup(context.TODO(), &iam.GetGroupInput{
		GroupName: &group,
	})

	if err != nil {
		return nil, err
	}

	for _, user := range resp.Users {
		users = append(users, User{Account: *user.UserName, Name: strings.ReplaceAll(*user.UserName, ".", "")})
	}

	log.Debug().Msgf("IAM users: %v", users)
	return users, nil
}

func (awsSdkCfg *AwsSdkConfig) SaveConfS3(env string, user string, filePath string) (string, error) {
	s3Client := s3.NewFromConfig(awsSdkCfg.SdkConfig)
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	uploader := manager.NewUploader(s3Client)
	key := fmt.Sprintf("%s/%s.ovpn", env, user)
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: &awsSdkCfg.AwsConfig.BucketName,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	// Generate a presigned URL for the uploaded file in S3
	presign := s3.NewPresignClient(s3Client)
	req, err := presign.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &awsSdkCfg.AwsConfig.BucketName,
		Key:    &key},
		s3.WithPresignExpires(sessionDuration))
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("Presign url generated for env %s user %s: %s", env, user, req.URL)

	return req.URL, nil
}

func (awsSdkCfg *AwsSdkConfig) SendMail(env string, user User, urlStr string, senderMail string) error {
	subject := fmt.Sprintf("Your VPN access to %s", env)
	recipient, _ := awsSdkCfg.GetEmail(user.Account)
	sender := senderMail

	message := mailTXT + urlStr

	emailInput := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{recipient},
		},
		Message: &types.Message{
			Body: &types.Body{
				Text: &types.Content{
					Data: &message,
				},
			},
			Subject: &types.Content{
				Data: &subject,
			},
		},
		Source: &sender,
	}

	sesClient := ses.NewFromConfig(awsSdkCfg.SdkConfig)
	_, err := sesClient.SendEmail(context.TODO(), emailInput)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Email sent with the presigned URL for env %s : %s", env, user.Name)
	return nil
}

func (awsSdkCfg *AwsSdkConfig) RemoveConfS3(env string, user string) error {
	key := fmt.Sprintf("%s/%s.ovpn", env, user)
	s3Client := s3.NewFromConfig(awsSdkCfg.SdkConfig)
	input := &s3.DeleteObjectInput{
		Bucket: &awsSdkCfg.AwsConfig.BucketName,
		Key:    &key,
	}
	_, err := s3Client.DeleteObject(context.TODO(), input)
	if err != nil {
		return err
	}
	return nil
}

func (awsSdkCfg *AwsSdkConfig) GetEmail(user string) (string, error) {
	iamClient := iam.NewFromConfig(awsSdkCfg.SdkConfig)
	input := &iam.ListUserTagsInput{
		UserName: &user,
	}

	result, err := iamClient.ListUserTags(context.TODO(), input)
	if err != nil {
		return "", err
	}

	var emailTagValue string
	for _, tag := range result.Tags {
		if *tag.Key == "email" {
			emailTagValue = *tag.Value
			break
		}
	}

	if emailTagValue == "" {
		return "", fmt.Errorf("email tag not found for IAM user: %s", user)
	}

	return emailTagValue, nil
}
