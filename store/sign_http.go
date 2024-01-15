package store

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"

	"github.com/if-itb/siasn-libs-backend/httputil"
)

// HandleActivityAdmissionSubmit handles a new admission request.
// Activity admission is created when an agency want to hold an event for their civil servants. A new activity
// will be created in the database with status set to `newly admitted`.
func (c *Client) HandleSignSubmit(w http.ResponseWriter, r *http.Request) {

	urlPath := "https://esign.bkn.go.id/api/sign/pdf"
	method := "POST"
	nik := r.FormValue("nik")
	passpharase := r.FormValue("passpharase")
	tag_koordinat := r.FormValue("tag_koordinat")
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	// orang.NomorIDDokumen = "3275080903800022"
	writer.WriteField("nik", nik)
	writer.WriteField("passphrase", passpharase)
	writer.WriteField("image", "true")
	writer.WriteField("width", "100")
	writer.WriteField("tampilan", "visible")
	writer.WriteField("height", "20")
	writer.WriteField("tag_koordinat", tag_koordinat)
	writer.WriteField("jenis_response", "BASE64")

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	file, err := os.Open(currentDir + "\\uploads\\" + "doc1.pdf")
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filePdf := make(textproto.MIMEHeader)
	filePdf.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, file.Name()))
	filePdf.Set("Content-Type", "application/pdf")
	fileWriterPdf, err := writer.CreatePart(filePdf)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	_, err = io.Copy(fileWriterPdf, file)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fileImage, err := os.Open(currentDir + "\\uploads\\" + "logo-bsre.png")
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filePng := make(textproto.MIMEHeader)
	filePng.Set("Content-Disposition", fmt.Sprintf(`form-data; name="imageTTD"; filename="%s"`, fileImage.Name()))
	filePng.Set("Content-Type", "image/png")
	filewriterPng, err := writer.CreatePart(filePng)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	_, err = io.Copy(filewriterPng, fileImage)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Close the form writer
	err = writer.Close()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	writer.Close()
	client := &http.Client{}
	req, err := http.NewRequest(method, urlPath, payload)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	req.Header.Add("Authorization", "Basic YXBpX3NpYXNuMjE6U0lBU05AQktOMjAyMQ==")
	req.Header.Add("Cookie", "643b92a03b9ae4752f4e2b34c409b950=a725a7008d3dfffa7c623e49ce666b59")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	_ = httputil.WriteObj200HtmlEscape(w, map[string]string{
		"response": string(body),
	}, false)
}
