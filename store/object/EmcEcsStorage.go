package object

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/EMCECS/ecs-object-client-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fazrithe/siasn-jf-backend-git/libs/logutil"
	"github.com/google/uuid"
)

type EmcEcsStorage struct {
	Client *ecs.S3
	Logger logutil.Logger
	// TempBucket represents the bucket name to store temporary files.
	TempBucket string
	// TempActivityDir represents a directory to store temporary activity files.
	// It does not start or end with a slash. It is relative to TempBucket, so files will be stored in
	// TempBucket/TempActivityDir/filename.
	TempActivityDir string
	// ActivityBucket represents the bucket name to store activity files (support documents).
	ActivityBucket string
	// ActivityDir represents a directory to store activity support documents.
	// It does not start or end with a slash. It is relative to ActivityBucket, so files will be stored in
	// ActivityBucket/ActivityDir/filename.
	ActivityDir string

	// TempRequirementDir represents a directory to store temporary requirement files.
	// It does not start or end with a slash. It is relative to TempBucket, so files will be stored in
	// TempBucket/TempRequirementDir/filename.
	TempRequirementDir string
	// RequirementBucket represents the bucket name to store requirement files (calculation documents, cover letters, etc.).
	RequirementBucket string
	// RequirementDir represents a directory to store requirement documents.
	// It does not start or end with a slash. It is relative to RequirementBucket, so files will be stored in
	// RequirementBucket/RequirementDir/filename.
	RequirementDir string
	// RequirementTemplateDir represents directory for storing template documents.
	RequirementTemplateDir string
	// RequirementTemplateCoverLetterFilename represents the filename of cover letter template including its extension.
	RequirementTemplateCoverLetterFilename string

	// TempDismissalDir represents a directory to store temporary Dismissal files.
	// It does not start or end with a slash. It is relative to TempBucket, so files will be stored in
	// TempBucket/TempDismissalDir/filename.
	TempDismissalDir string
	// DismissalBucket represents the bucket name to store Dismissal files (calculation documents, cover letters, etc.).
	DismissalBucket string
	// DismissalDir represents a directory to store Dismissal documents.
	// It does not start or end with a slash. It is relative to DismissalBucket, so files will be stored in
	// DismissalBucket/DismissalDir/filename.
	DismissalDir string
	// DismissalTemplateDir represents directory for storing dismissal template documents.
	DismissalTemplateDir string
	// DismissalTemplateAcceptanceLetterFilename represents the filename of dismissal acceptance letter template including its extension.
	DismissalTemplateAcceptanceLetterFilename string

	// TempPromotionDir represents a directory to store temporary Promotion files.
	// It does not start or end with a slash. It is relative to TempBucket, so files will be stored in
	// TempBucket/TempPromotionDir/filename.
	TempPromotionDir string
	// PromotionBucket represents the bucket name to store Promotion files (calculation documents, cover letters, etc.).
	PromotionBucket string
	// PromotionDir represents a directory to store Promotion documents.
	// It does not start or end with a slash. It is relative to PromotionBucket, so files will be stored in
	// PromotionBucket/PromotionDir/filename.
	PromotionDir string
	// PromotionTemplateDir represents directory for storing promotion template documents.
	PromotionTemplateDir string
	// PromotionTemplatePakLetterFilename represents the filename of promotion pak letter template including its extension.
	PromotionTemplatePakLetterFilename string
	// PromotionTemplateRecommendationLetterFilename represents the filename of promotion recommendation letter template including its extension.
	PromotionTemplateRecommendationLetterFilename string

	// PromotionCpnsBucket represents the bucket name to store Promotion for CPNS files.
	PromotionCpnsBucket string
	// PromotionCpnsDir represents a directory to store Promotion for CPNS documents.
	// It does not start or end with a slash. It is relative to PromotionCpnsBucket, so files will be stored in
	// PromotionCpnsBucket/PromotionDir/filename.
	PromotionCpnsDir string
	// TempPromotionCpnsDir represents a directory to store temporary Promotion for CPNS files.
	// It does not start or end with a slash. It is relative to PromotionCpnsBucket, so files will be stored in
	// PromotionCpnsBucket/TempPromotionCpnsDir/filename.
	TempPromotionCpnsDir string

	// TempAssessmentTeamDir represents a directory to store temporary assessment team files.
	// It does not start or end with a slash. It is relative to TempBucket, so files will be stored in
	// TempBucket/TempAssessmentTeamDir/filename.
	TempAssessmentTeamDir string
	// AssessmentTeamBucket represents the bucket name to store assessment team files (support documents).
	AssessmentTeamBucket string
	// AssessmentTeamDir represents a directory to store assessment team support documents.
	// It does not start or end with a slash. It is relative to AssessmentTeamBucket, so files will be stored in
	// AssessmentTeamBucket/AssessmentTeamDir/filename.
	AssessmentTeamDir string

	// SignUrlExpire is the duration in which signed URL will expire after being generated, for any purposes.
	SignUrlExpire time.Duration
}

func (s *EmcEcsStorage) getFileMetadata(ctx context.Context, bucket, filename string) (metadata *Metadata, err error) {
	out, err := s.Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		if er, ok := err.(awserr.Error); ok {
			if er.Code() == "NoSuchKey" || er.Code() == "NotFound" {
				return nil, ErrFileNotFound
			}
		}
		return nil, err
	}

	return &Metadata{
		Bucket:        bucket,
		Filename:      filename,
		Checksum:      aws.StringValue(out.ETag),
		ContentType:   aws.StringValue(out.ContentType),
		ContentLength: int(aws.Int64Value(out.ContentLength)),
		LastModified:  aws.TimeValue(out.LastModified),
	}, nil
}

// getFile retrieves a file from the bucket.
func (s *EmcEcsStorage) getFile(ctx context.Context, bucket, filename string) (out io.ReadCloser, err error) {
	awsOut, err := s.Client.GetObjectWithContext(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filename),
		},
	)
	if err != nil {
		if er, ok := err.(awserr.Error); ok {
			if er.Code() == "NoSuchKey" || er.Code() == "NotFound" {
				return nil, ErrFileNotFound
			}
		}
		return nil, err
	}

	return awsOut.Body, nil
}

// putFile puts a file into the bucket.
func (s *EmcEcsStorage) putFile(ctx context.Context, bucket, filename string, contentType string, in io.ReadSeeker) (err error) {
	_, err = s.Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:        in,
		Bucket:      aws.String(bucket),
		ContentType: aws.String(contentType),
		Key:         aws.String(filename),
	})
	return err
}

// saveFiles moves files from temporary bucket to permanent one.
func (s *EmcEcsStorage) saveFiles(ctx context.Context, tempBucket, tempDir, permanentBucket, permanentDir string, filenames []string, deleteOriginal bool) (results []*SaveResult, err error) {
	if len(filenames) >= 1000 {
		return nil, ErrTooManyFiles
	}

	var tempObjects = make([]*s3.ObjectIdentifier, 0)

	for _, filename := range filenames {
		tempObjects = append(tempObjects, &s3.ObjectIdentifier{Key: aws.String(path.Join(tempDir, filename))})
		var out *s3.CopyObjectOutput
		out, err = s.Client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
			Bucket:     aws.String(permanentBucket),
			CopySource: aws.String(path.Join(tempBucket, tempDir, filename)),
			Key:        aws.String(path.Join(permanentDir, filename)),
		})
		if err != nil {
			if er, ok := err.(awserr.Error); ok {
				if er.Code() == "NoSuchKey" || er.Code() == "NotFound" {
					return nil, ErrTempFileNotFound
				}
			}
			return nil, err
		}

		if out.CopyObjectResult == nil || out.CopyObjectResult.ETag == nil || out.CopyObjectResult.LastModified == nil {
			return nil, errors.New("copied object does not have E-Tag or Last-Modified")
		}
		result := &SaveResult{
			Bucket:    permanentBucket,
			Dir:       permanentDir,
			Filename:  filename,
			Checksum:  *out.CopyObjectResult.ETag,
			CreatedAt: *out.CopyObjectResult.LastModified,
		}
		results = append(results, result)
	}

	if deleteOriginal {
		_, err = s.Client.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{Bucket: aws.String(tempBucket), Delete: &s3.Delete{Objects: tempObjects}})
		if err != nil {
			s.Logger.Warnf("cannot delete temporary files, skipping: %v", err)
		}
	}

	return results, nil
}

func (s *EmcEcsStorage) generateGetSignedUrl(ctx context.Context, bucket string, key string, expireDuration time.Duration) (signedUrl *url.URL, err error) {
	req, _ := s.Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	req.SetContext(ctx)
	urlStr, err := req.Presign(expireDuration)
	return url.Parse(urlStr)
}

func (s *EmcEcsStorage) GetActivityFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	meta, err := s.getFileMetadata(ctx, s.ActivityBucket, path.Join(s.ActivityDir, filename))
	if err != nil {
		return nil, err
	}

	meta.Dir = s.ActivityDir
	return meta, nil
}

func (s *EmcEcsStorage) GetActivityFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	return s.getFile(ctx, s.ActivityBucket, path.Join(s.ActivityDir, filename))
}

func (s *EmcEcsStorage) PutActivityFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return s.putFile(ctx, s.ActivityBucket, path.Join(s.ActivityDir, filename), contentType, in)
}

func (s *EmcEcsStorage) SaveActivityFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error) {
	return s.saveFiles(ctx, s.TempBucket, s.TempActivityDir, s.ActivityBucket, s.ActivityDir, filenames, true)
}

func (s *EmcEcsStorage) GenerateActivityFilename(mimeType string) (filename string, err error) {
	ext, ok := MimeTypeToExtension[mimeType]
	if !ok {
		return "", ErrFileTypeUnsupported
	}

	return fmt.Sprintf("%s.%s", uuid.New().String(), ext), nil
}

func (s *EmcEcsStorage) GenerateActivityDocPutSign(_ context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.TempBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.TempActivityDir, filename)),
	})
	urlStr, err := req.Presign(s.SignUrlExpire)
	url, _ = url.Parse(urlStr)
	return url, err
}

func (s *EmcEcsStorage) GenerateActivityDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.ActivityBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.ActivityDir, filename)),
	})
	req.SetContext(ctx)
	urlStr, err := req.Presign(s.SignUrlExpire)
	if err != nil {
		return nil, err
	}
	return url.Parse(urlStr)
}

func (s *EmcEcsStorage) GenerateActivityDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.TempBucket, fmt.Sprintf("%s/%s", s.TempActivityDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) GenerateActivityDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.ActivityBucket, fmt.Sprintf("%s/%s", s.ActivityDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) GetRequirementFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	meta, err := s.getFileMetadata(ctx, s.RequirementBucket, path.Join(s.RequirementDir, filename))
	if err != nil {
		return nil, err
	}

	meta.Dir = s.RequirementDir
	return meta, nil
}

func (s *EmcEcsStorage) GetRequirementFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	return s.getFile(ctx, s.RequirementBucket, path.Join(s.RequirementDir, filename))
}

func (s *EmcEcsStorage) PutRequirementFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return s.putFile(ctx, s.RequirementBucket, path.Join(s.RequirementDir, filename), contentType, in)
}

func (s *EmcEcsStorage) SaveRequirementFile(ctx context.Context, src string, dest string) (result *SaveResult, err error) {
	var out *s3.CopyObjectOutput
	out, err = s.Client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.RequirementBucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s/%s", s.TempBucket, s.TempRequirementDir, src)),
		Key:        aws.String(fmt.Sprintf("%s/%s", s.RequirementDir, dest)),
	})
	if err != nil {
		if er, ok := err.(awserr.Error); ok {
			if er.Code() == "NoSuchKey" || er.Code() == "NotFound" {
				return nil, ErrTempFileNotFound
			}
		}
		return nil, err
	}

	if out.CopyObjectResult == nil || out.CopyObjectResult.ETag == nil || out.CopyObjectResult.LastModified == nil {
		return nil, errors.New("copied object does not have E-Tag or Last-Modified")
	}
	result = &SaveResult{
		Bucket:    s.RequirementBucket,
		Dir:       s.RequirementDir,
		Filename:  dest,
		Checksum:  *out.CopyObjectResult.ETag,
		CreatedAt: *out.CopyObjectResult.LastModified,
	}

	_, err = s.Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.TempBucket), Key: aws.String(fmt.Sprintf("%s/%s", s.TempRequirementDir, src))})
	if err != nil {
		s.Logger.Warnf("cannot delete temporary files, skipping: %v", err)
	}

	return result, nil
}

func (s *EmcEcsStorage) SaveRequirementFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error) {
	return s.saveFiles(ctx, s.TempBucket, s.TempRequirementDir, s.RequirementBucket, s.RequirementDir, filenames, true)
}

func (s *EmcEcsStorage) GenerateRequirementFilename(mimeType string) (filename string, err error) {
	return s.GenerateActivityFilename(mimeType)
}

func (s *EmcEcsStorage) GenerateRequirementDocPutSign(ctx context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.TempBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.TempRequirementDir, filename)),
	})
	req.SetContext(ctx)
	urlStr, err := req.Presign(s.SignUrlExpire)
	if err != nil {
		return nil, err
	}
	return url.Parse(urlStr)
}

func (s *EmcEcsStorage) GenerateRequirementDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.RequirementBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.RequirementDir, filename)),
	})
	req.SetContext(ctx)
	urlStr, err := req.Presign(s.SignUrlExpire)
	if err != nil {
		return nil, err
	}
	return url.Parse(urlStr)
}

func (s *EmcEcsStorage) GenerateRequirementDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.TempBucket, fmt.Sprintf("%s/%s", s.TempRequirementDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) GenerateRequirementDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.RequirementBucket, fmt.Sprintf("%s/%s", s.RequirementDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) GenerateRequirementTemplateCoverLetterGetSign(ctx context.Context) (url *url.URL, err error) {
	return s.generateGetSignedUrl(
		ctx,
		s.RequirementBucket,
		fmt.Sprintf("%s/%s", s.RequirementTemplateDir, s.RequirementTemplateCoverLetterFilename),
		s.SignUrlExpire,
	)
}

func (s *EmcEcsStorage) GetDismissalFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	meta, err := s.getFileMetadata(ctx, s.DismissalBucket, path.Join(s.DismissalDir, filename))
	if err != nil {
		return nil, err
	}

	meta.Dir = s.DismissalDir
	return meta, nil
}

func (s *EmcEcsStorage) GetDismissalFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	return s.getFile(ctx, s.DismissalBucket, path.Join(s.DismissalDir, filename))
}

func (s *EmcEcsStorage) PutDismissalFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return s.putFile(ctx, s.DismissalBucket, path.Join(s.DismissalDir, filename), contentType, in)
}

func (s *EmcEcsStorage) SaveDismissalFile(ctx context.Context, src string, dest string) (result *SaveResult, err error) {
	var out *s3.CopyObjectOutput
	out, err = s.Client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.DismissalBucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s/%s", s.TempBucket, s.TempDismissalDir, src)),
		Key:        aws.String(fmt.Sprintf("%s/%s", s.DismissalDir, dest)),
	})
	if err != nil {
		if er, ok := err.(awserr.Error); ok {
			if er.Code() == "NoSuchKey" || er.Code() == "NotFound" {
				return nil, ErrTempFileNotFound
			}
		}
		return nil, err
	}

	if out.CopyObjectResult == nil || out.CopyObjectResult.ETag == nil || out.CopyObjectResult.LastModified == nil {
		return nil, errors.New("copied object does not have E-Tag or Last-Modified")
	}
	result = &SaveResult{
		Bucket:    s.DismissalBucket,
		Dir:       s.DismissalDir,
		Filename:  dest,
		Checksum:  *out.CopyObjectResult.ETag,
		CreatedAt: *out.CopyObjectResult.LastModified,
	}

	_, err = s.Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.TempBucket), Key: aws.String(fmt.Sprintf("%s/%s", s.TempDismissalDir, src))})
	if err != nil {
		s.Logger.Warnf("cannot delete temporary files, skipping: %v", err)
	}

	return result, nil
}

func (s *EmcEcsStorage) SaveDismissalFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error) {
	return s.saveFiles(ctx, s.TempBucket, s.TempDismissalDir, s.DismissalBucket, s.DismissalDir, filenames, true)
}

func (s *EmcEcsStorage) GenerateDismissalDocName(mimeType string) (filename string, err error) {
	return s.GenerateActivityFilename(mimeType)
}

func (s *EmcEcsStorage) GenerateDismissalDocPutSign(ctx context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.TempBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.TempDismissalDir, filename)),
	})
	req.SetContext(ctx)
	urlStr, err := req.Presign(s.SignUrlExpire)
	if err != nil {
		return nil, err
	}
	return url.Parse(urlStr)
}

func (s *EmcEcsStorage) GenerateDismissalDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.DismissalBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.DismissalDir, filename)),
	})
	req.SetContext(ctx)
	urlStr, err := req.Presign(s.SignUrlExpire)
	if err != nil {
		return nil, err
	}
	return url.Parse(urlStr)
}

func (s *EmcEcsStorage) GenerateDismissalDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.TempBucket, fmt.Sprintf("%s/%s", s.TempDismissalDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) GenerateDismissalDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.DismissalBucket, fmt.Sprintf("%s/%s", s.DismissalDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) GenerateDismissalTemplateAcceptanceLetterGetSign(ctx context.Context) (url *url.URL, err error) {
	return s.generateGetSignedUrl(
		ctx,
		s.DismissalBucket,
		fmt.Sprintf("%s/%s", s.DismissalTemplateDir, s.DismissalTemplateAcceptanceLetterFilename),
		s.SignUrlExpire,
	)
}

func (s *EmcEcsStorage) GetPromotionFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	meta, err := s.getFileMetadata(ctx, s.PromotionBucket, path.Join(s.PromotionDir, filename))
	if err != nil {
		return nil, err
	}

	meta.Dir = s.PromotionDir
	return meta, nil
}

func (s *EmcEcsStorage) GetPromotionFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	return s.getFile(ctx, s.PromotionBucket, path.Join(s.PromotionDir, filename))
}

func (s *EmcEcsStorage) PutPromotionFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return s.putFile(ctx, s.PromotionBucket, path.Join(s.PromotionDir, filename), contentType, in)
}

func (s *EmcEcsStorage) SavePromotionFile(ctx context.Context, src string, dest string) (result *SaveResult, err error) {
	var out *s3.CopyObjectOutput
	out, err = s.Client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.PromotionBucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s/%s", s.TempBucket, s.TempPromotionDir, src)),
		Key:        aws.String(fmt.Sprintf("%s/%s", s.PromotionDir, dest)),
	})
	if err != nil {
		if er, ok := err.(awserr.Error); ok {
			if er.Code() == "NoSuchKey" || er.Code() == "NotFound" {
				return nil, ErrTempFileNotFound
			}
		}
		return nil, err
	}

	if out.CopyObjectResult == nil || out.CopyObjectResult.ETag == nil || out.CopyObjectResult.LastModified == nil {
		return nil, errors.New("copied object does not have E-Tag or Last-Modified")
	}
	result = &SaveResult{
		Bucket:    s.PromotionBucket,
		Dir:       s.PromotionDir,
		Filename:  dest,
		Checksum:  *out.CopyObjectResult.ETag,
		CreatedAt: *out.CopyObjectResult.LastModified,
	}

	_, err = s.Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.TempBucket), Key: aws.String(fmt.Sprintf("%s/%s", s.TempPromotionDir, src))})
	if err != nil {
		s.Logger.Warnf("cannot delete temporary files, skipping: %v", err)
	}

	return result, nil
}

func (s *EmcEcsStorage) SavePromotionFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error) {
	return s.saveFiles(ctx, s.TempBucket, s.TempPromotionDir, s.PromotionBucket, s.PromotionDir, filenames, true)
}

func (s *EmcEcsStorage) GeneratePromotionDocName(mimeType string) (filename string, err error) {
	return s.GenerateActivityFilename(mimeType)
}

func (s *EmcEcsStorage) GeneratePromotionDocPutSign(ctx context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.TempBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.TempPromotionDir, filename)),
	})
	req.SetContext(ctx)
	urlStr, err := req.Presign(s.SignUrlExpire)
	if err != nil {
		return nil, err
	}
	return url.Parse(urlStr)
}

func (s *EmcEcsStorage) GeneratePromotionDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.PromotionBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.PromotionDir, filename)),
	})
	req.SetContext(ctx)
	urlStr, err := req.Presign(s.SignUrlExpire)
	if err != nil {
		return nil, err
	}
	return url.Parse(urlStr)
}

func (s *EmcEcsStorage) GeneratePromotionDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.TempBucket, fmt.Sprintf("%s/%s", s.TempPromotionDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) GeneratePromotionDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.PromotionBucket, fmt.Sprintf("%s/%s", s.PromotionDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) GeneratePromotionTemplatePakLetterGetSign(ctx context.Context) (url *url.URL, err error) {
	return s.generateGetSignedUrl(
		ctx,
		s.PromotionBucket,
		fmt.Sprintf("%s/%s", s.PromotionTemplateDir, s.PromotionTemplatePakLetterFilename),
		s.SignUrlExpire,
	)
}

func (s *EmcEcsStorage) GeneratePromotionTemplateRecommendationLetterGetSign(ctx context.Context) (url *url.URL, err error) {
	return s.generateGetSignedUrl(
		ctx,
		s.PromotionBucket,
		fmt.Sprintf("%s/%s", s.PromotionTemplateDir, s.PromotionTemplateRecommendationLetterFilename),
		s.SignUrlExpire,
	)
}

func (s *EmcEcsStorage) GetPromotionCpnsFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	meta, err := s.getFileMetadata(ctx, s.PromotionCpnsBucket, path.Join(s.PromotionCpnsDir, filename))
	if err != nil {
		return nil, err
	}

	meta.Dir = s.PromotionCpnsDir
	return meta, nil
}

func (s *EmcEcsStorage) GetPromotionCpnsFileMetadataTemp(ctx context.Context, filename string) (metadata *Metadata, err error) {
	meta, err := s.getFileMetadata(ctx, s.TempBucket, path.Join(s.PromotionCpnsDir, filename))
	if err != nil {
		return nil, err
	}

	meta.Dir = s.PromotionCpnsDir
	return meta, nil
}

func (s *EmcEcsStorage) GetPromotionCpnsFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	return s.getFile(ctx, s.PromotionCpnsBucket, path.Join(s.PromotionCpnsDir, filename))
}

func (s *EmcEcsStorage) PutPromotionCpnsFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return s.putFile(ctx, s.PromotionCpnsBucket, path.Join(s.PromotionCpnsDir, filename), contentType, in)
}

func (s *EmcEcsStorage) SavePromotionCpnsFile(ctx context.Context, src string, dest string, deleteOriginal bool) (result *SaveResult, err error) {
	var out *s3.CopyObjectOutput
	out, err = s.Client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.PromotionCpnsBucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s/%s", s.TempBucket, s.TempPromotionCpnsDir, src)),
		Key:        aws.String(fmt.Sprintf("%s/%s", s.PromotionCpnsDir, dest)),
	})
	if err != nil {
		if er, ok := err.(awserr.Error); ok {
			if er.Code() == "NoSuchKey" || er.Code() == "NotFound" {
				return nil, ErrTempFileNotFound
			}
		}
		return nil, err
	}

	if out.CopyObjectResult == nil || out.CopyObjectResult.ETag == nil || out.CopyObjectResult.LastModified == nil {
		return nil, errors.New("copied object does not have E-Tag or Last-Modified")
	}
	result = &SaveResult{
		Bucket:    s.PromotionCpnsBucket,
		Dir:       s.PromotionCpnsDir,
		Filename:  dest,
		Checksum:  *out.CopyObjectResult.ETag,
		CreatedAt: *out.CopyObjectResult.LastModified,
	}

	if deleteOriginal {
		_, err = s.Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.TempBucket), Key: aws.String(fmt.Sprintf("%s/%s", s.TempPromotionDir, src))})
		if err != nil {
			s.Logger.Warnf("cannot delete temporary files, skipping: %v", err)
		}
	}

	return result, nil
}

func (s *EmcEcsStorage) SavePromotionCpnsFiles(ctx context.Context, filenames []string, deleteOriginal bool) (results []*SaveResult, err error) {
	return s.saveFiles(ctx, s.TempBucket, s.TempPromotionCpnsDir, s.PromotionCpnsBucket, s.PromotionCpnsDir, filenames, deleteOriginal)
}

func (s *EmcEcsStorage) GeneratePromotionCpnsDocName(mimeType string) (filename string, err error) {
	return s.GenerateActivityFilename(mimeType)
}

func (s *EmcEcsStorage) GeneratePromotionCpnsDocPutSign(ctx context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.TempBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.TempPromotionCpnsDir, filename)),
	})
	req.SetContext(ctx)
	urlStr, err := req.Presign(s.SignUrlExpire)
	if err != nil {
		return nil, err
	}
	return url.Parse(urlStr)
}

func (s *EmcEcsStorage) GeneratePromotionCpnsDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.TempBucket, fmt.Sprintf("%s/%s", s.TempPromotionCpnsDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) GeneratePromotionCpnsDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.PromotionCpnsBucket, fmt.Sprintf("%s/%s", s.PromotionCpnsDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) DeletePromotionCpnsFilesTemp(ctx context.Context, filenames []string) (err error) {
	tempObjects := make([]*s3.ObjectIdentifier, 0)
	for _, f := range filenames {
		tempObjects = append(tempObjects, &s3.ObjectIdentifier{Key: aws.String(path.Join(s.TempPromotionCpnsDir, f))})
	}

	_, err = s.Client.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{Bucket: aws.String(s.TempBucket), Delete: &s3.Delete{Objects: tempObjects}})
	return err
}

func (s *EmcEcsStorage) GetAssessmentTeamFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	meta, err := s.getFileMetadata(ctx, s.AssessmentTeamBucket, path.Join(s.AssessmentTeamDir, filename))
	if err != nil {
		return nil, err
	}

	meta.Dir = s.AssessmentTeamDir
	return meta, nil
}

func (s *EmcEcsStorage) GetAssessmentTeamFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	return s.getFile(ctx, s.AssessmentTeamBucket, path.Join(s.AssessmentTeamDir, filename))
}

func (s *EmcEcsStorage) PutAssessmentTeamFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return s.putFile(ctx, s.AssessmentTeamBucket, path.Join(s.AssessmentTeamDir, filename), contentType, in)
}

func (s *EmcEcsStorage) SaveAssessmentTeamFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error) {
	return s.saveFiles(ctx, s.TempBucket, s.TempAssessmentTeamDir, s.AssessmentTeamBucket, s.AssessmentTeamDir, filenames, true)
}

func (s *EmcEcsStorage) GenerateAssessmentTeamFilename(mimeType string) (filename string, err error) {
	ext, ok := MimeTypeToExtension[mimeType]
	if !ok {
		return "", ErrFileTypeUnsupported
	}

	return fmt.Sprintf("%s.%s", uuid.New().String(), ext), nil
}

func (s *EmcEcsStorage) GenerateAssessmentTeamDocPutSign(_ context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.TempBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.TempAssessmentTeamDir, filename)),
	})
	urlStr, err := req.Presign(s.SignUrlExpire)
	url, _ = url.Parse(urlStr)
	return url, err
}

func (s *EmcEcsStorage) GenerateAssessmentTeamDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.AssessmentTeamBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.AssessmentTeamDir, filename)),
	})
	req.SetContext(ctx)
	urlStr, err := req.Presign(s.SignUrlExpire)
	if err != nil {
		return nil, err
	}
	return url.Parse(urlStr)
}

func (s *EmcEcsStorage) GenerateAssessmentTeamDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.TempBucket, fmt.Sprintf("%s/%s", s.TempAssessmentTeamDir, filename), s.SignUrlExpire)
}

func (s *EmcEcsStorage) GenerateAssessmentTeamDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return s.generateGetSignedUrl(ctx, s.AssessmentTeamBucket, fmt.Sprintf("%s/%s", s.AssessmentTeamDir, filename), s.SignUrlExpire)
}
