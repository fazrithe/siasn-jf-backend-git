// Package object holds all the functionalities needed to interact with the object storage.
// All functions in this package does not support error code implemented by the standard ec and errnum package.
package object

import (
	"context"
	"errors"
	"io"
	"net/url"
	"time"
)

var ErrTooManyFiles = errors.New("number of files exceed limit")

// Deprecated: renamed (replaced Admission with Activity).
var ErrTempAdmissionFileNotFound = ErrTempFileNotFound
var ErrTempFileNotFound = errors.New("cannot copy temporary admission file, file not found in temporary bucket")
var ErrFileNotFound = errors.New("file not found in the bucket")

// Deprecated: renamed (replaced Admission with Activity).
var ErrAdmissionFileTypeUnsupported = ErrFileTypeUnsupported
var ErrFileTypeUnsupported = errors.New("mime type is not supported for admission support doc file, support only: " +
	"application/pdf, application/vnd.openxmlformats-officedocument.spreadsheetml.sheet, application/vnd.ms-excel, application/vnd.openxmlformats-officedocument.wordprocessingml.document, application/msword")

type Storage interface {
	ActivityStorage
	RequirementStorage
	DismissalStorage
	PromotionStorage
	PromotionCpnsStorage
	AssessmentTeamStorage
}

type ActivityStorage interface {
	// GetActivityFileMetadata retrieves a file metadata (HEAD) from files in permanent bucket.
	GetActivityFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error)

	// GetActivityFile retrieves a file from permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	GetActivityFile(ctx context.Context, filename string) (out io.ReadCloser, err error)

	// PutActivityFile puts a file to permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	PutActivityFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error)

	// SaveActivityFiles moves temporary activity files from temporary location to permanent location which
	// may reside in different bucket.
	//
	// You can add a directory in the filename, but do not include a leading slash.
	// For example, you can use a filename like this: support/filename.jpg.
	SaveActivityFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error)

	// GenerateActivityFilename generates activity support document name by creating a UUID and add to it a
	// proper extension based on the given mime type. If the mime type is not supported, error is returned.
	// Only these are allowed:
	//   application/pdf => .pdf
	//   application/vnd.openxmlformats-officedocument.spreadsheetml.sheet => .xlsx
	//   application/vnd.ms-excel => .xls
	//   application/vnd.openxmlformats-officedocument.wordprocessingm => .docx
	//   application/msword => .doc
	GenerateActivityFilename(mimeType string) (filename string, err error)

	// GenerateActivityDocPutSign generates a signed URL to PUT an activity support doc to the temporary location.
	GenerateActivityDocPutSign(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateActivityDocPutSignDirect generates a signed URL to PUT an activity support doc directly to the permanent location.
	GenerateActivityDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateActivityDocGetSignTemp generates a signed URL to GET an activity support doc from the temporary location.
	GenerateActivityDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateActivityDocGetSign generates a signed URL to GET an activity support doc from the permanent location.
	GenerateActivityDocGetSign(ctx context.Context, filename string) (url *url.URL, err error)
}

type RequirementStorage interface {
	// GetRequirementFileMetadata retrieves a file metadata (HEAD) from files in permanent bucket.
	GetRequirementFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error)

	// GetRequirementFile retrieves a file from permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	GetRequirementFile(ctx context.Context, filename string) (out io.ReadCloser, err error)

	// PutRequirementFile puts a file to permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	PutRequirementFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error)

	// SaveRequirementFile moves a file from src to dest.
	// src and dest are both filenames without bucket name. This will save only a single file, and is useful to
	// move files such as cover letters that was stored in different name in temporary location. The file will be
	// copied into the destination with new name.
	SaveRequirementFile(ctx context.Context, src string, dest string) (result *SaveResult, err error)

	// SaveRequirementFiles moves temporary requirement files from temporary location to permanent location which
	// may reside in different bucket.
	//
	// You can add a directory in the filename, but do not include a leading slash.
	// For example, you can use a filename like this: support/filename.jpg.
	SaveRequirementFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error)

	// GenerateRequirementFilename generates activity support document name by creating a UUID and add to it a
	// proper extension based on the given mime type. If the mime type is not supported, error is returned.
	// Only these are allowed:
	//   application/pdf => .pdf
	//   application/vnd.openxmlformats-officedocument.spreadsheetml.sheet => .xlsx
	//   application/vnd.ms-excel => .xls
	//   application/vnd.openxmlformats-officedocument.wordprocessingm => .docx
	//   application/msword => .doc
	GenerateRequirementFilename(mimeType string) (filename string, err error)

	// GenerateRequirementDocPutSign generates a signed URL to PUT a requirement support doc to the temporary location.
	GenerateRequirementDocPutSign(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateRequirementDocPutSignDirect generates a signed URL to PUT a requirement support doc directly to the permanent location.
	GenerateRequirementDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateRequirementDocGetSignTemp generates a signed URL to GET a requirement support doc from the temporary location.
	GenerateRequirementDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateRequirementDocGetSign generates a signed URL to GET a requirement support doc from the permanent location.
	GenerateRequirementDocGetSign(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateRequirementTemplateCoverLetterGetSign generates a signed URL to GET the cover letter template.
	GenerateRequirementTemplateCoverLetterGetSign(ctx context.Context) (url *url.URL, err error)
}

type DismissalStorage interface {
	// GetDismissalFileMetadata retrieves a file metadata (HEAD) from files in permanent bucket.
	GetDismissalFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error)

	// GetDismissalFile retrieves a file from permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	GetDismissalFile(ctx context.Context, filename string) (out io.ReadCloser, err error)

	// PutDismissalFile puts a file to permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	PutDismissalFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error)

	// SaveDismissalFile moves a file from src to dest.
	// src and dest are both filenames without bucket name. This will save only a single file, and is useful to
	// move files such as cover letters that was stored in different name in temporary location. The file will be
	// copied into the destination with new name.
	SaveDismissalFile(ctx context.Context, src string, dest string) (result *SaveResult, err error)

	// SaveDismissalFiles moves temporary dismissal files from temporary location to permanent location which
	// may reside in different bucket.
	//
	// You can add a directory in the filename, but do not include a leading slash.
	// For example, you can use a filename like this: support/filename.jpg.
	SaveDismissalFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error)

	// GenerateDismissalDocName generates activity support document name by creating a UUID and add to it a
	// proper extension based on the given mime type. If the mime type is not supported, error is returned.
	// Only these are allowed:
	//   application/pdf => .pdf
	//   application/vnd.openxmlformats-officedocument.spreadsheetml.sheet => .xlsx
	//   application/vnd.ms-excel => .xls
	//   application/vnd.openxmlformats-officedocument.wordprocessingm => .docx
	//   application/msword => .doc
	GenerateDismissalDocName(mimeType string) (filename string, err error)

	// GenerateDismissalDocPutSign generates a signed URL to PUT a dismissal support doc to the temporary location.
	GenerateDismissalDocPutSign(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateDismissalDocPutSignDirect generates a signed URL to PUT a promotion document file directly to the permanent location.
	GenerateDismissalDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateDismissalDocGetSignTemp generates a signed URL to GET a dismissal support doc from the temporary location.
	GenerateDismissalDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateDismissalDocGetSign generates a signed URL to GET a dismissal support doc from the permanent location.
	GenerateDismissalDocGetSign(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateDismissalTemplateAcceptanceLetterGetSign generates a signed URL to GET the dismissal acceptance letter template.
	GenerateDismissalTemplateAcceptanceLetterGetSign(ctx context.Context) (url *url.URL, err error)
}

type PromotionStorage interface {
	// GetPromotionFileMetadata retrieves a file metadata (HEAD) from files in permanent bucket.
	GetPromotionFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error)

	// GetPromotionFile retrieves a file from permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	GetPromotionFile(ctx context.Context, filename string) (out io.ReadCloser, err error)

	// PutPromotionFile puts a file to permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	PutPromotionFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error)

	// SavePromotionFile moves a file from src to dest.
	// src and dest are both filenames without bucket name. This will save only a single file, and is useful to
	// move files such as cover letters that was stored in different name in temporary location. The file will be
	// copied into the destination with new name.
	SavePromotionFile(ctx context.Context, src string, dest string) (result *SaveResult, err error)

	// SavePromotionFiles moves temporary promotion files from temporary location to permanent location which
	// may reside in different bucket.
	//
	// You can add a directory in the filename, but do not include a leading slash.
	// For example, you can use a filename like this: support/filename.jpg.
	SavePromotionFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error)

	// GeneratePromotionDocName generates promotion document file name by creating a UUID and add to it a
	// proper extension based on the given mime type. If the mime type is not supported, error is returned.
	// Only these are allowed:
	//   application/pdf => .pdf
	//   application/vnd.openxmlformats-officedocument.spreadsheetml.sheet => .xlsx
	//   application/vnd.ms-excel => .xls
	//   application/vnd.openxmlformats-officedocument.wordprocessingm => .docx
	//   application/msword => .doc
	GeneratePromotionDocName(mimeType string) (filename string, err error)

	// GeneratePromotionDocPutSign generates a signed URL to PUT a promotion document file to the temporary location.
	GeneratePromotionDocPutSign(ctx context.Context, filename string) (url *url.URL, err error)

	// GeneratePromotionDocPutSignDirect generates a signed URL to PUT a promotion document file directly to the permanent location.
	GeneratePromotionDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error)

	// GeneratePromotionDocGetSignTemp generates a signed URL to GET a promotion document file from the temporary location.
	GeneratePromotionDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error)

	// GeneratePromotionDocGetSign generates a signed URL to GET a promotion document file from the permanent location.
	GeneratePromotionDocGetSign(ctx context.Context, filename string) (url *url.URL, err error)

	// GeneratePromotionTemplatePakLetterGetSign generates a signed URL to GET the promotion pak letter template.
	GeneratePromotionTemplatePakLetterGetSign(ctx context.Context) (url *url.URL, err error)

	// GeneratePromotionTemplateRecommendationLetterGetSign generates a signed URL to GET the promotion recommendation letter template.
	GeneratePromotionTemplateRecommendationLetterGetSign(ctx context.Context) (url *url.URL, err error)
}

type PromotionCpnsStorage interface {
	// GetPromotionCpnsFileMetadata retrieves a file metadata (HEAD) from files in permanent bucket.
	GetPromotionCpnsFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error)

	// GetPromotionCpnsFileMetadataTemp retrieves a file metadata (HEAD) from files in temporary bucket.
	GetPromotionCpnsFileMetadataTemp(ctx context.Context, filename string) (metadata *Metadata, err error)

	// GetPromotionCpnsFile retrieves a file from permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	GetPromotionCpnsFile(ctx context.Context, filename string) (out io.ReadCloser, err error)

	// PutPromotionCpnsFile puts a file to permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	PutPromotionCpnsFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error)

	// SavePromotionCpnsFile moves a file from src to dest.
	// src and dest are both filenames without bucket name. This will save only a single file, and is useful to
	// move files such as cover letters that was stored in different name in temporary location. The file will be
	// copied into the destination with new name.
	SavePromotionCpnsFile(ctx context.Context, src string, dest string, deleteOriginal bool) (result *SaveResult, err error)

	// SavePromotionCpnsFiles moves temporary promotion files from temporary location to permanent location which
	// may reside in different bucket.
	//
	// You can add a directory in the filename, but do not include a leading slash.
	// For example, you can use a filename like this: support/filename.jpg.
	SavePromotionCpnsFiles(ctx context.Context, filenames []string, deleteOriginal bool) (results []*SaveResult, err error)

	// GeneratePromotionCpnsDocName generates cpns promotion document file name by creating a UUID and add to it a
	// proper extension based on the given mime type. If the mime type is not supported, error is returned.
	// Only these are allowed:
	//   application/pdf => .pdf
	//   application/vnd.openxmlformats-officedocument.spreadsheetml.sheet => .xlsx
	//   application/vnd.ms-excel => .xls
	//   application/vnd.openxmlformats-officedocument.wordprocessingm => .docx
	//   application/msword => .doc
	GeneratePromotionCpnsDocName(mimeType string) (filename string, err error)

	// GeneratePromotionCpnsDocPutSign generates a signed URL to PUT a document file for CPNS promotion to the temporary location.
	GeneratePromotionCpnsDocPutSign(ctx context.Context, filename string) (url *url.URL, err error)

	// 	GeneratePromotionCpnsDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error) generates a signed URL to GET a promotion document file from the temporary location.
	GeneratePromotionCpnsDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error)

	// GeneratePromotionCpnsDocGetSign generates a signed URL to GET a promotion document file from the permanent location.
	GeneratePromotionCpnsDocGetSign(ctx context.Context, filename string) (url *url.URL, err error)

	// DeletePromotionCpnsFilesTemp deletes files from temporary location.
	DeletePromotionCpnsFilesTemp(ctx context.Context, filenames []string) (err error)
}

type AssessmentTeamStorage interface {
	// GetAssessmentTeamFileMetadata retrieves a file metadata (HEAD) from files in permanent bucket.
	GetAssessmentTeamFileMetadata(ctx context.Context, filename string) (metadata *Metadata, err error)

	// GetAssessmentTeamFile retrieves a file from permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	GetAssessmentTeamFile(ctx context.Context, filename string) (out io.ReadCloser, err error)

	// PutAssessmentTeamFile puts a file to permanent bucket.
	// In the future this might be deprecated as all modules have combined bucket, or the interface gets generalized.
	PutAssessmentTeamFile(ctx context.Context, filename string, contentType string, in io.ReadSeeker) (err error)

	// SaveAssessmentTeamFiles moves temporary assessment team files from temporary location to permanent location which
	// may reside in different bucket.
	//
	// You can add a directory in the filename, but do not include a leading slash.
	// For example, you can use a filename like this: support/filename.jpg.
	SaveAssessmentTeamFiles(ctx context.Context, filenames []string) (results []*SaveResult, err error)

	// GenerateAssessmentTeamFilename generates assessment team support document name by creating a UUID and add to it a
	// proper extension based on the given mime type. If the mime type is not supported, error is returned.
	// Only these are allowed:
	//   application/pdf => .pdf
	//   application/vnd.openxmlformats-officedocument.spreadsheetml.sheet => .xlsx
	//   application/vnd.ms-excel => .xls
	//   application/vnd.openxmlformats-officedocument.wordprocessingm => .docx
	//   application/msword => .doc
	GenerateAssessmentTeamFilename(mimeType string) (filename string, err error)

	// GenerateAssessmentTeamDocPutSign generates a signed URL to PUT an assessment team support doc to the temporary location.
	GenerateAssessmentTeamDocPutSign(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateAssessmentTeamDocPutSignDirect generates a signed URL to PUT an assessment team support doc directly to the permanent location.
	GenerateAssessmentTeamDocPutSignDirect(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateAssessmentTeamDocGetSignTemp generates a signed URL to GET an assessment team support doc from the temporary location.
	GenerateAssessmentTeamDocGetSignTemp(ctx context.Context, filename string) (url *url.URL, err error)

	// GenerateAssessmentTeamDocGetSign generates a signed URL to GET an assessment team support doc from the permanent location.
	GenerateAssessmentTeamDocGetSign(ctx context.Context, filename string) (url *url.URL, err error)
}

// Deprecated: renamed (replaced Admission with Activity).
var AdmissionMimeTypeToExtension = MimeTypeToExtension

// MimeTypeToExtension stores mime type to filename extension (without dot) mappings
// supported for uploading various documents.
var MimeTypeToExtension = map[string]string{
	"application/pdf": "pdf",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "xlsx",
	"application/vnd.ms-excel": "xls",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "docx",
	"application/msword": "doc",
}

// SaveResult is a struct containing the result of saving a document.
// It holds the document filename, checksum, and created at timestamp. The returned filename is a canonical filename.
// The directory is also returned in this struct. The object key therefore can be derived from dir + filename.
type SaveResult struct {
	Bucket    string
	Dir       string
	Filename  string
	Checksum  string
	CreatedAt time.Time
}

type Metadata struct {
	Bucket        string
	Dir           string
	Filename      string
	Checksum      string
	ContentType   string
	LastModified  time.Time
	ContentLength int
}
