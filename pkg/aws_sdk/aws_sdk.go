package aws_sdk

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
)

//func GetObjectList(bucketName string) []types.Object {
//	// Создаем кастомный обработчик эндпоинтов, который для сервиса S3 и региона ru-central1 выдаст корректный URL
//	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
//		if service == s3.ServiceID && region == "ru-central1" {
//			return aws.Endpoint{
//				PartitionID:   "yc",
//				URL:           "https://storage.yandexcloud.net",
//				SigningRegion: "ru-central1",
//			}, nil
//		}
//		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
//	})
//	// Подгружаем конфигрурацию из ~/.aws/*
//	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(customResolver))
//	if err != nil {
//		logrus.Fatal(err, 1)
//	}
//
//	// Создаем клиента для доступа к хранилищу S3
//	client := s3.NewFromConfig(cfg)
//
//	// Запрашиваем список бакетов
//	result, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
//		Bucket: aws.String(bucketName),
//	})
//	if err != nil {
//		logrus.Fatal(err, 2)
//	}
//
//	return result.Contents
//}

func GetObjectFromYandexCloud(bucketName, key string) ([]byte, error) {
	accessKey := os.Getenv("AWS_ACCESS_KEY")
	secretKey := os.Getenv("AWS_SECRET_KEY")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
			},
		}),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "yc",
				URL:           "https://storage.yandexcloud.net",
				SigningRegion: "ru-central1",
			}, nil
		})),
	)

	if err != nil {
		return []byte{}, errors.Errorf("error while loading AWS config: %s", err.Error())
	}

	client := s3.NewFromConfig(cfg)

	resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &key,
	})
	if err != nil {
		return []byte{}, errors.Errorf("error while trying load object: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logrus.Errorf("memory leak: error closing response body: %s", err.Error())
		}
	}(resp.Body)

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, errors.Errorf("error while reading resp.body: %s", err.Error())
	}

	return bytes, nil
}

// TODO create a function which creating bucket or folder and function which put object into folder

func PutObjectInBucket(bucketName, key string, body *os.File) error {
	accessKey := os.Getenv("AWS_ACCESS_KEY")
	secretKey := os.Getenv("AWS_SECRET_KEY")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
			},
		}),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "yc",
				URL:           "https://storage.yandexcloud.net",
				SigningRegion: "ru-central1",
			}, nil
		})),
	)

	if err != nil {
		return errors.Errorf("error while loading AWS config: %s", err.Error())
	}

	client := s3.NewFromConfig(cfg)

	params := &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &key,
		Body:   body,
	}

	outputParams, err := client.PutObject(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Println(*outputParams)
	if err := body.Close(); err != nil {
		fmt.Println(err)
	}
	return nil

	//if err != nil {
	//	return []byte{}, errors.Errorf("error while trying load object: %s", err.Error())
	//}
	//defer func(Body io.ReadCloser) {
	//	err := Body.Close()
	//	if err != nil {
	//		logrus.Errorf("memory leak: error closing response body: %s", err.Error())
	//	}
	//}(resp.Body)
	//
	//bytes, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return []byte{}, errors.Errorf("error while reading resp.body: %s", err.Error())
	//}

	//return bytes, nil
}
