package docx

import (
	"context"
)

// Renderer provides the capability to render docx documents cluttered with jinja2 templates into a docx document
// or pdf, given a set of data. The default implementation uses siasn-docx script to render docx template into docx
// document and uses soffice command from LibreOffice to render docx document into pdf.
type Renderer interface {
	// Render renders a template located with templatePath into another docx document saved in outputPath.
	// It should overwrite existing file in outputPath. Will not throw any error if data does not contain any
	// value that can be found in the template. Template values not found in data will just be empty.
	Render(data interface{}, templatePath string, outputPath string) (err error)
	// RenderCtx is Render that supports context.
	RenderCtx(ctx context.Context, data interface{}, templatePath string, outputPath string) (err error)
	// RenderAsPdf works just like Render but will also convert the resulting docx document into pdf.
	// Output path here, therefore, should be a .pdf file.
	RenderAsPdf(data interface{}, templatePath string, outputPath string) (err error)
	// RenderAsPdfCtx is RenderAsPdf that supports context.
	RenderAsPdfCtx(ctx context.Context, data interface{}, templatePath string, outputPath string) (err error)
}
