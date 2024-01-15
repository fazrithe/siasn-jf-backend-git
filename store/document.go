package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/if-itb/siasn-jf-backend/errnum"
	"github.com/if-itb/siasn-jf-backend/store/models"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/httputil"
)

const (
	TimeoutActivityDocumentUpload = TimeoutDefault
)

func (c *Client) HandleActivityDocumentSubmit(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(10 << 20)           // 10 MB limit
	file, handler, err := r.FormFile("file") // "file" is the name of the file input field in the HTML form
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	defer file.Close()
	fileExt := filepath.Ext(handler.Filename)
	originalFileName := strings.TrimSuffix(filepath.Base(handler.Filename), filepath.Ext(handler.Filename))
	now := time.Now()
	filename := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt
	out, err := os.Create("uploads/" + filename)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		log.Fatal(err)
	}
	// Save the file to a local directory
	// f, err := os.OpenFile("uploads/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	// if err != nil {
	// 	fmt.Fprintf(w, "Error: %v", err)
	// 	return
	// }
	// defer f.Close()
	// io.Copy(f, file)
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSubmit)
	defer cancel()
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return
	}
	defer func() {
		c.completeMtx(mtx, err)
	}()
	id := uuid.New()
	name := r.FormValue("name")
	modul := r.FormValue("modul")
	penandatangan := r.FormValue("penandatangan")
	penandatanganArray := []string{penandatangan}
	penandatanganJson, err := json.MarshalIndent(penandatanganArray, "", "  ")
	penandatanganString := string(penandatanganJson)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	_, err = mtx.Exec("insert into dokumen_template(id, modul, filename, penandatangan, name) values($1, $2, $3, $4, $5)",
		id,
		modul,
		filename,
		penandatanganString,
		name,
	)
	if err != nil {
		fmt.Fprintf(w, "Error inserting into database: %v", err)
		return
	}
	if err != nil {
		return
	}
	_ = httputil.WriteObj200HtmlEscape(w, map[string]string{
		"Message": "Insert Success",
	}, false)

}

func (c *Client) HandleActivityDocumentGet(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSubmit)
	defer cancel()

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return
	}
	rows, err := mtx.Query("SELECT id, name, modul, filename, penandatangan FROM dokumen_template")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var data []models.Dokumen_tamplate_item
	for rows.Next() {
		var d models.Dokumen_tamplate_item
		if err := rows.Scan(&d.ID, &d.Name, &d.Modul, &d.Filename, &d.Penandatangan); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data = append(data, d)

	}

	// if d.Filename == "" {
	// 	c.httpError(w, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
	// 	return
	// }
	// url, err := c.AssessmentTeamStorage.GenerateAssessmentTeamDocGetSignTemp(ctx, path.Join(AssessmentTeamSupportDocSubdir, "piring2-1698307920.jpg"))
	// if err != nil {
	// 	c.httpError(w, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
	// 	return
	// }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (c *Client) HandleActivityDocumentDownload(writer http.ResponseWriter, request *http.Request) {
	type schemaFilename struct {
		Filename string `schema:"filename"`
	}

	s := &schemaFilename{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.Filename == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	// ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityCertGenDocDownload)
	// defer cancel()

	// url, err := c.ActivityStorage.GenerateActivityDocGetSign(ctx, path.Join(ActivitySupportDocSubdir, s.Filename))
	// if err != nil {
	// 	c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
	// 	return
	// }

	// http.Redirect(writer, request, url.String(), http.StatusFound)

	// _ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
	// 	"Message": "Success",
	// 	"URL":     url.String(),
	// }, false)
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	f, err := os.Open(currentDir + "\\uploads\\" + s.Filename)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	contentDisposition := fmt.Sprintf("attachment; filename=%s", f.Name())
	writer.Header().Set("Content-Disposition", contentDisposition)

	if _, err := io.Copy(writer, f); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Client) HandleActivityDocumentDelete(writer http.ResponseWriter, request *http.Request) {
	type schemaDocumentId struct {
		ID string `schema:"id"`
	}

	s := &schemaDocumentId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.ID == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSubmit)
	defer cancel()
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return
	}
	defer func() {
		c.completeMtx(mtx, err)
	}()
	_, err = mtx.ExecContext(ctx, "delete from dokumen_template where id = $1", s.ID)
	if err != nil {
		fmt.Fprintf(writer, "Error delete: %v", err)
		return
	}
	if err != nil {
		return
	}
	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"Message": "Delete Success",
	}, false)
}
