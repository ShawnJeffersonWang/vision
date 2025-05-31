package alioss

import (
	"fmt"
	"mime/multipart"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"agricultural_vision/settings"
)

// 初始化oss服务
func InitServer() (*oss.Client, error) {
	endpoint := settings.Conf.AliossConfig.Endpoint
	accessKeyId := settings.Conf.AliossConfig.AccessKeyId
	accessKeySecret := settings.Conf.AliossConfig.AccessKeySecret

	// 创建OSSClient实例。
	ossClient, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		return nil, err
	}
	return ossClient, nil
}

// 上传文件
func UploadFile(file multipart.File, fileName, pathName string) (fileURL string, err error) {
	client, err := InitServer()
	if err != nil {
		return
	}

	// 获取 Bucket
	bucket, err := client.Bucket(settings.Conf.AliossConfig.BucketName)
	if err != nil {
		return
	}

	// OSS 完整路径：path + 文件名
	objectName := pathName + fileName

	// 上传文件
	err = bucket.PutObject(objectName, file)
	if err != nil {
		return
	}

	// 返回文件访问 URL
	fileURL = fmt.Sprintf("https://%s.%s/%s",
		settings.Conf.AliossConfig.BucketName,
		settings.Conf.AliossConfig.Endpoint,
		objectName,
	)

	return
}
