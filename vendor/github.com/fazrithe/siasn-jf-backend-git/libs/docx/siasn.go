package docx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var (
	ErrSiasnRendererSaveTemplateToDocx = errors.New("cannot save template to docx")
	ErrSiasnRendererLoadTemplate       = errors.New("cannot load template")
	ErrSiasnRendererBadTemplate        = errors.New("bad syntax in template")
)

const (
	siasnErrCodeBadTemplate = 5
)

// SiasnRenderer works with siasn-docx script and soffice command.
// It works by encoding data into JSON first, to be able to be read by the siasn-docx command. The given
// render data can thus implement the json.Marshaler interface.
//
// Render data can also contain specialized object fields. They are listed in object.go file. For example
// InlineImage, which is used to render an image.
type SiasnRenderer struct {
	// The siasn-docx command path.
	// Can be full path or just command name, if it exists in PATH.
	DocxCmd string
	// Additional run arguments to be passed to siasn-docx command at the end of the default arguments.
	DocxArgs []string
	// The soffice command path.
	// Can be full path or just command name, if it exists in PATH.
	SofficeCmd string
	// Additional run arguments to be passed to soffice command at the end of the default arguments.
	SofficeArgs []string
}

// NewSiasnRenderer creates a default SiasnRenderer with siasn-docx and soffice command assumed
// to be installed and can be called in PATH.
//
// To create SiasnRenderer with custom command paths, create it manually.
//
// Will panic if templatePath/outputPath is empty or invalid.
func NewSiasnRenderer() *SiasnRenderer {
	return &SiasnRenderer{
		DocxCmd:    "siasn-docx",
		SofficeCmd: "soffice",
	}
}

type siasnError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *SiasnRenderer) Render(data interface{}, templatePath string, outputPath string) (err error) {
	return s.RenderCtx(context.Background(), data, templatePath, outputPath)
}

func (s *SiasnRenderer) RenderCtx(ctx context.Context, data interface{}, templatePath string, outputPath string) (err error) {
	if templatePath == "" {
		panic("templatePath cannot be nil")
	}

	if outputPath == "" {
		panic("outputPath cannot be nil")
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	args := []string{"--json", string(payload)}
	args = append(args, s.DocxArgs...)
	args = append(args, path.Clean(templatePath), path.Clean(outputPath))

	stderr := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, s.DocxCmd, args...)
	cmd.Stderr = stderr
	out, err := cmd.Output()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			siasnErr := &siasnError{}
			err = json.Unmarshal(out, siasnErr)
			if err != nil {
				rawOut := out
				if out == nil || len(out) == 0 {
					rawOut = stderr.Bytes()
				}
				return fmt.Errorf("cannot decode output message: %w, output: %s, exit code: %d", err, string(rawOut), exitErr.ExitCode())
			}

			switch siasnErr.Code {
			case 3:
				return fmt.Errorf("%w: %d: %s", ErrSiasnRendererSaveTemplateToDocx, siasnErr.Code, siasnErr.Message)
			case 4:
				return fmt.Errorf("%w: %d: %s", ErrSiasnRendererLoadTemplate, siasnErr.Code, siasnErr.Message)
			case siasnErrCodeBadTemplate:
				return fmt.Errorf("%w: %d: %s", ErrSiasnRendererBadTemplate, siasnErr.Code, siasnErr.Message)
			default:
				return fmt.Errorf("%d: %s", siasnErr.Code, siasnErr.Message)
			}
		}

		return err
	}

	return nil
}

func (s *SiasnRenderer) RenderAsPdf(data interface{}, templatePath string, outputPath string) (err error) {
	return s.RenderAsPdfCtx(context.Background(), data, templatePath, outputPath)
}

func (s *SiasnRenderer) RenderAsPdfCtx(ctx context.Context, data interface{}, templatePath string, outputPath string) (err error) {
	tmpDir := os.TempDir()
	tmpOutputBasename := fmt.Sprintf("compiled-%s", path.Base(templatePath))
	tmpOutputPath := path.Join(tmpDir, tmpOutputBasename)
	// Try to delete tmp output regardless if it exists or not
	defer os.Remove(tmpOutputPath)

	err = s.RenderCtx(ctx, data, templatePath, tmpOutputPath)
	if err != nil {
		return err
	}

	tmpOutputBasenameNoExt := strings.TrimSuffix(tmpOutputBasename, filepath.Ext(tmpOutputBasename))
	pdfTmpOutputPath := path.Join(tmpDir, fmt.Sprintf("%s.pdf", tmpOutputBasenameNoExt))

	args := []string{"--convert-to", "pdf", "--headless", "--outdir", tmpDir, tmpOutputPath}
	args = append(args, s.SofficeArgs...)
	cmd := exec.CommandContext(ctx, s.SofficeCmd, args...)
	err = cmd.Run()
	if err != nil {
		return err
	}

	cleanOutputPath := path.Clean(outputPath)
	err = os.Rename(pdfTmpOutputPath, cleanOutputPath)
	if err != nil {
		linkError := &os.LinkError{}
		if errors.As(err, &linkError) { // Retry with manual copying
			err = s.manualMoveFile(cleanOutputPath, pdfTmpOutputPath)
			if err != nil {
				return err
			}
		}
		return err
	}

	return nil
}

// manualMoveFile does a manual file move/rename by copying the file and deleting the source.
// It is used when os.Rename cannot be used, for example, moving between partitions.
func (s *SiasnRenderer) manualMoveFile(dst string, src string) (err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer os.Remove(src)
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buffer := make([]byte, 4*1024)
	_, err = io.CopyBuffer(dstFile, srcFile, buffer)
	if err != nil {
		return err
	}

	return nil
}
