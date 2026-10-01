package remover

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ObjectVersion は列挙・削除対象の1オブジェクトを表す。
type ObjectVersion struct {
	Key            string
	VersionId      string
	IsDeleteMarker bool
}

// Remover はS3操作のコアロジックを持つ構造体。
type Remover struct {
	client *s3.Client
	bucket string
	prefix string
}

// NewRemover はAWS SDK v2クライアントを初期化してRemoverを返す。
// regionが空の場合はAWS_REGION環境変数または~/.aws/configのデフォルト設定を使用する。
func NewRemover(ctx context.Context, bucket, prefix, region string) (*Remover, error) {
	opts := []func(*config.LoadOptions) error{}
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	return &Remover{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}, nil
}

// CheckBucketExists はHeadBucket APIでバケットの存在を確認する。
// バケットが存在しない場合はエラーを返す。
func (r *Remover) CheckBucketExists(ctx context.Context) error {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucket),
	})
	if err != nil {
		return fmt.Errorf("bucket %q not found: %w", r.bucket, err)
	}
	return nil
}

// ListObjects はListObjectVersions APIで指定プレフィックスのオブジェクトを全件列挙する。
// 通常バージョン・過去バージョン・削除マーカーをすべて含む。
func (r *Remover) ListObjects(ctx context.Context) ([]ObjectVersion, error) {
	var objects []ObjectVersion
	var keyMarker *string
	var versionIdMarker *string

	for {
		input := &s3.ListObjectVersionsInput{
			Bucket: aws.String(r.bucket),
			Prefix: aws.String(r.prefix),
		}
		if keyMarker != nil {
			input.KeyMarker = keyMarker
		}
		if versionIdMarker != nil {
			input.VersionIdMarker = versionIdMarker
		}

		output, err := r.client.ListObjectVersions(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, v := range output.Versions {
			objects = append(objects, ObjectVersion{
				Key:            aws.ToString(v.Key),
				VersionId:      aws.ToString(v.VersionId),
				IsDeleteMarker: false,
			})
		}

		for _, d := range output.DeleteMarkers {
			objects = append(objects, ObjectVersion{
				Key:            aws.ToString(d.Key),
				VersionId:      aws.ToString(d.VersionId),
				IsDeleteMarker: true,
			})
		}

		if !aws.ToBool(output.IsTruncated) {
			break
		}
		keyMarker = output.NextKeyMarker
		versionIdMarker = output.NextVersionIdMarker
	}

	return objects, nil
}

// DeleteObjects はバッチ削除（最大1000件/リクエスト）でオブジェクトを削除する。
// 削除に失敗したオブジェクトを返す。
func (r *Remover) DeleteObjects(ctx context.Context, objects []ObjectVersion) ([]ObjectVersion, error) {
	var failed []ObjectVersion

	for _, chunk := range chunkObjectVersions(objects, 1000) {
		identifiers := make([]types.ObjectIdentifier, len(chunk))
		for i, obj := range chunk {
			identifiers[i] = types.ObjectIdentifier{
				Key:       aws.String(obj.Key),
				VersionId: aws.String(obj.VersionId),
			}
		}

		output, err := r.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(r.bucket),
			Delete: &types.Delete{
				Objects: identifiers,
				Quiet:   aws.Bool(false),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete objects: %w", err)
		}

		for _, e := range output.Errors {
			failed = append(failed, ObjectVersion{
				Key:       aws.ToString(e.Key),
				VersionId: aws.ToString(e.VersionId),
			})
		}
	}

	return failed, nil
}

// chunkObjectVersions はスライスをsize件ずつのチャンクに分割する。
func chunkObjectVersions(objects []ObjectVersion, size int) [][]ObjectVersion {
	if len(objects) == 0 {
		return [][]ObjectVersion{}
	}

	var chunks [][]ObjectVersion
	for i := 0; i < len(objects); i += size {
		end := i + size
		if end > len(objects) {
			end = len(objects)
		}
		chunks = append(chunks, objects[i:end])
	}
	return chunks
}
