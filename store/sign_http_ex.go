package store

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func (c *Client) HandleSignTte(w http.ResponseWriter, r *http.Request) {

	url := "https://esign.bkn.go.id/api/sign/pdf"

	// Prepare the file to be uploaded
	// currentDir, err := os.Getwd()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }
	filePath := "D:/BKN/proyek/siasn-jf-backend/uploads/doc1.pdf"
	f1, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(f1)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
	// str1 := string(f1.Name())
	fmt.Println("Image:", f1.Name())
	// Prepare the form data
	formData := map[string]string{
		"height":         "40",
		"image":          "true",
		"jenis_response": "BASE64",
		"nik":            "3275080903800022",
		"passphrase":     "P4ssword2023!",
		"tag_koordinat":  "$",
		"tampilan":       "VISIBLE",
		"width":          "300",
	}
	// filePath := "D:/BKN/proyek/siasn-jf-backend/uploads/doc1.pdf"

	// Make the request
	response, err := makePostRequest1(url, formData)
	if err != nil {
		fmt.Println("Error making the request:", err)
		return
	}

	defer response.Body.Close()

	// Read the response body
	responseBody, err := readResponseBody1(response)
	if err != nil {
		fmt.Println("Error reading the response body:", err)
		return
	}

	// Print the response
	fmt.Println("Response:", string(responseBody))
}

func makePostRequest1(url string, formData map[string]string) (*http.Response, error) {
	// Create a buffer to store the form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	for key, value := range formData {
		writer.WriteField(key, value)
	}

	// Add file field
	var fileBuffer bytes.Buffer
	// currentDir, err := os.Getwd()
	filePath := "/D:/BKN/proyek/siasn-jf-backend/uploads/doc1.pdf"
	filePath2 := "D:/BKN/proyek/siasn-jf-backend/uploads/doc1.pdf"
	file, err := os.Open(filePath2)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Copy the file content to the buffer
	_, err = io.Copy(&fileBuffer, file)
	if err != nil {
		log.Fatal(err)
	}
	filePart, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(filePart, &fileBuffer)
	if err != nil {
		log.Fatal(err)
	}

	// Close the multipart writer
	writer.Close()
	// Create the request
	contentDisposition := fmt.Sprintf("attachment; file=%s", file.Name())
	// writer.Header().Set("Content-Disposition", contentDisposition)

	req, err := http.NewRequest("POST", url, body)
	username := "api_siasn21"
	password := "SIASN@BKN2021"
	auth := username + ":" + password
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Disposition", contentDisposition)

	if err != nil {
		return nil, err
	}

	// Set the content type for form data
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make the request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func readResponseBody1(response *http.Response) ([]byte, error) {
	// Read the response body
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(response.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
