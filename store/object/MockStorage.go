package object

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/url"
	"time"
)

// MockStorage implements the Storage interface, but will never return any error.
type MockStorage struct{}

func (m *MockStorage) GetActivityFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	return &Metadata{
		Bucket:        "",
		Dir:           "",
		Filename:      filename,
		Checksum:      filename,
		ContentType:   "application/pdf",
		LastModified:  time.Now(),
		ContentLength: int(time.Now().UnixNano()),
	}, nil
}

func (m *MockStorage) GetActivityFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	buffer := bytes.NewBufferString(uuid.NewString())
	return ioutil.NopCloser(buffer), nil
}

func (m *MockStorage) PutActivityFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return nil
}

func (m *MockStorage) GetRequirementFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	return &Metadata{
		Bucket:        "",
		Dir:           "",
		Filename:      filename,
		Checksum:      filename,
		ContentType:   "application/pdf",
		LastModified:  time.Now(),
		ContentLength: int(time.Now().UnixNano()),
	}, nil
}

func (m *MockStorage) GetRequirementFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	buffer := bytes.NewBufferString(uuid.NewString())
	return ioutil.NopCloser(buffer), nil
}

func (m *MockStorage) PutRequirementFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return nil
}

func (m *MockStorage) GetDismissalFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	return &Metadata{
		Bucket:        "",
		Dir:           "",
		Filename:      filename,
		Checksum:      filename,
		ContentType:   "application/pdf",
		LastModified:  time.Now(),
		ContentLength: int(time.Now().UnixNano()),
	}, nil
}

func (m *MockStorage) GetDismissalFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	buffer := bytes.NewBufferString(uuid.NewString())
	return ioutil.NopCloser(buffer), nil
}

func (m *MockStorage) PutDismissalFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return nil
}

func (m *MockStorage) GetPromotionFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	return &Metadata{
		Bucket:        "",
		Dir:           "",
		Filename:      filename,
		Checksum:      filename,
		ContentType:   "application/pdf",
		LastModified:  time.Now(),
		ContentLength: int(time.Now().UnixNano()),
	}, nil
}

func (m *MockStorage) GetPromotionFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	buffer := bytes.NewBufferString(uuid.NewString())
	return ioutil.NopCloser(buffer), nil
}

func (m *MockStorage) PutPromotionFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return nil
}

func (m *MockStorage) SavePromotionFile(ctx context.Context, src string, dest string) (result *SaveResult, err error) {
	return m.SaveRequirementFile(ctx, src, dest)
}

func (m *MockStorage) SavePromotionFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error) {
	return m.SaveRequirementFiles(ctx, filenames)
}

func (m *MockStorage) GeneratePromotionDocName(mimeType string) (filename string, err error) {
	return uuid.New().String(), nil
}

func (m *MockStorage) GeneratePromotionDocPutSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GeneratePromotionDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GeneratePromotionDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GeneratePromotionDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GeneratePromotionTemplatePakLetterGetSign(ctx context.Context) (url *url.URL, err error) {
	return url.Parse("https://google.com/")
}

func (m *MockStorage) GeneratePromotionTemplateRecommendationLetterGetSign(ctx context.Context) (url *url.URL, err error) {
	return url.Parse("https://google.com/")
}

func (m *MockStorage) SaveDismissalFile(ctx context.Context, src string, dest string) (result *SaveResult, err error) {
	return m.SaveRequirementFile(ctx, src, dest)
}

func (m *MockStorage) SaveDismissalFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error) {
	return m.SaveRequirementFiles(ctx, filenames)
}

func (m *MockStorage) GenerateDismissalDocName(mimeType string) (filename string, err error) {
	return uuid.New().String(), nil
}

func (m *MockStorage) GenerateDismissalDocPutSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateDismissalDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateDismissalDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateDismissalDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) SaveActivityFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error) {
	bucket := uuid.New().String()
	dir := uuid.New().String()
	for _, filename := range filenames {
		results = append(results, &SaveResult{
			Bucket:    bucket,
			Dir:       dir,
			Filename:  filename,
			Checksum:  uuid.New().String(),
			CreatedAt: time.Now(),
		})
	}
	return results, nil
}

func (m *MockStorage) GenerateActivityFilename(mimeType string) (filename string, err error) {
	return uuid.New().String(), nil
}

func (m *MockStorage) GenerateActivityDocPutSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateActivityDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateActivityDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateActivityDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) SaveRequirementFile(ctx context.Context, src string, dest string) (result *SaveResult, err error) {
	return &SaveResult{
		Bucket:    "",
		Dir:       "",
		Filename:  dest,
		Checksum:  uuid.New().String(),
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockStorage) SaveRequirementFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error) {
	return m.SaveActivityFiles(ctx, filenames)
}

func (m *MockStorage) GenerateRequirementFilename(mimeType string) (filename string, err error) {
	return uuid.New().String(), nil
}

func (m *MockStorage) GenerateRequirementDocPutSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateRequirementDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateRequirementDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateRequirementDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateRequirementTemplateCoverLetterGetSign(ctx context.Context) (url *url.URL, err error) {
	return url.Parse("https://google.com/")
}

func (m *MockStorage) GenerateDismissalTemplateAcceptanceLetterGetSign(ctx context.Context) (url *url.URL, err error) {
	return url.Parse("https://google.com/")
}

func (m *MockStorage) GetPromotionCpnsFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	return &Metadata{
		Bucket:        "",
		Dir:           "",
		Filename:      filename,
		Checksum:      filename,
		ContentType:   "application/pdf",
		LastModified:  time.Now(),
		ContentLength: int(time.Now().UnixNano()),
	}, nil
}

func (m *MockStorage) GetPromotionCpnsFileMetadataTemp(ctx context.Context, filename string) (metadata *Metadata, err error) {
	return &Metadata{
		Bucket:        "",
		Dir:           "",
		Filename:      filename,
		Checksum:      filename,
		ContentType:   "application/pdf",
		LastModified:  time.Now(),
		ContentLength: int(time.Now().UnixNano()),
	}, nil
}

func (m *MockStorage) GetPromotionCpnsFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	return m.GetActivityFile(ctx, filename)
}

func (m *MockStorage) PutPromotionCpnsFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return nil
}

func (m *MockStorage) SavePromotionCpnsFile(ctx context.Context, src string, dest string, deleteOriginal bool) (result *SaveResult, err error) {
	return m.SaveRequirementFile(ctx, src, dest)
}

func (m *MockStorage) SavePromotionCpnsFiles(ctx context.Context, filenames []string, deleteOriginal bool) (results []*SaveResult, err error) {
	return m.SaveActivityFiles(ctx, filenames)
}

func (m *MockStorage) GeneratePromotionCpnsDocName(mimeType string) (filename string, err error) {
	return uuid.New().String(), nil
}

func (m *MockStorage) GeneratePromotionCpnsDocPutSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GeneratePromotionCpnsDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GeneratePromotionCpnsDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) DeletePromotionCpnsFilesTemp(ctx context.Context, filenames []string) (err error) {
	return nil
}

func (m *MockStorage) GetAssessmentTeamFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error) {
	return &Metadata{
		Bucket:        "",
		Dir:           "",
		Filename:      filename,
		Checksum:      filename,
		ContentType:   "application/pdf",
		LastModified:  time.Now(),
		ContentLength: int(time.Now().UnixNano()),
	}, nil
}

func (m *MockStorage) GetAssessmentTeamFile(ctx context.Context, filename string) (out io.ReadCloser, err error) {
	buffer := bytes.NewBufferString(uuid.NewString())
	return ioutil.NopCloser(buffer), nil
}

func (m *MockStorage) PutAssessmentTeamFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error) {
	return nil
}

func (m *MockStorage) SaveAssessmentTeamFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error) {
	bucket := uuid.New().String()
	dir := uuid.New().String()
	for _, filename := range filenames {
		results = append(results, &SaveResult{
			Bucket:    bucket,
			Dir:       dir,
			Filename:  filename,
			Checksum:  uuid.New().String(),
			CreatedAt: time.Now(),
		})
	}
	return results, nil
}

func (m *MockStorage) GenerateAssessmentTeamFilename(mimeType string) (filename string, err error) {
	return uuid.New().String(), nil
}

func (m *MockStorage) GenerateAssessmentTeamDocPutSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateAssessmentTeamDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateAssessmentTeamDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}

func (m *MockStorage) GenerateAssessmentTeamDocGetSign(ctx context.Context, filename string) (url *url.URL, err error) {
	return url.Parse("https://google.com/" + filename)
}
