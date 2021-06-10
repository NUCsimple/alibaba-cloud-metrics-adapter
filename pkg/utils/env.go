package utils

import (
	"errors"
	"os"
)

func GetRegionFromEnv() (region string, err error) {
	region = os.Getenv("Region")
	if region == "" {
		return "", errors.New("not found region info in env")
	}
	return region, nil
}

func GetAccessKeyIdFromEnv() (region string, err error) {
	region = os.Getenv("AccessKeyId")
	if region == "" {
		return "", errors.New("not found ak in env")
	}
	return region, nil
}

func GetAccessKeySecretFromEnv() (region string, err error) {
	region = os.Getenv("AccessKeySecret")
	if region == "" {
		return "", errors.New("not found sk in env")
	}
	return region, nil
}
