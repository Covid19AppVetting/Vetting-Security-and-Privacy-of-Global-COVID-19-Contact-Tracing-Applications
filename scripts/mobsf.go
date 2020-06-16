package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

//const API_KEY = "8aabd97fc1fa10b5f60ad710145a58a3d6a32f1b3d5f7bc04a3354c56fb6bd5f"
const API_KEY = "045f9b8dfc25ea2e78df5c90ea0383413a97fb2a179cb873f269af2ecc8cd558"
const BASE_URL = "http://localhost:8000/"

const BASE_PATH = "/Users/zachwang/Develop/assignments/MPhil/Contact-Tracing-Apps/Contact Tracing Apks"
const SAVING_PATH = "/data/ApkReport"

var quit chan int = make(chan int, 4)
var wg = sync.WaitGroup{}

type MyFile struct {
	path string
	filename string
}

//upload and generate pdf
func main() {

	var files []MyFile
	err := filepath.Walk(BASE_PATH,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			filename := info.Name()
			if !info.IsDir() && filename[len(filename)-3:len(filename)] == "apk" {

				files = append(files, MyFile{
					path:     path,
					filename: filename,
				})
			}
			return nil
		})

	if err != nil {
		log.Println(err)
	}
	jobsCount := len(files)
	for i := 0; i < jobsCount; i++ {
		file := files[i]
		wg.Add(1)
		go func(filepath string, filename string) {
			quit <- i
			uploadApk(filepath, filename)
			<-quit
		}(file.path, file.filename)
	}

	for i := 0; i < jobsCount; i++ {
		fmt.Printf("index: %d,goroutine Num: %d\n", i, runtime.NumGoroutine())
	}


	wg.Wait()
	close(quit)
	fmt.Println("all done")
}

func uploadApk(filepath string, filename string){

	defer wg.Done()

	contentByte := SendPostFileRequest(BASE_URL + "api/v1/upload", filepath)
	var objmap map[string]string
	err := json.Unmarshal(contentByte, &objmap)
	if err != nil || objmap["error"] != ""{
		fmt.Println(objmap["error"])
		return
	}

	fmt.Println("analyzing", filename)
	_ = SendPostRequest(BASE_URL + "api/v1/scan", objmap)

	savingPath := SAVING_PATH + filepath[len(BASE_PATH):len(filepath)-3] + "pdf"

	savingParentDir := SAVING_PATH + filepath[len(BASE_PATH):len(filepath)-len(filename)]

	if _, err := os.Stat(savingPath); os.IsExist(err) {
		_ = os.Remove(savingPath)
	}

	if _, err := os.Stat(savingParentDir); os.IsNotExist(err) {
		_ = os.MkdirAll(savingParentDir, 0777)
	}

	_ = SendDownloadRequest(BASE_URL + "api/v1/download_pdf", map[string]string{"hash": objmap["hash"]}, savingPath)
	fmt.Println("done", filename)
}

func SendPostFileRequest(url string, filename string) []byte {
	file, err := os.Open(filename)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()


	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))

	if err != nil {
		log.Fatal(err)
	}

	io.Copy(part, file)
	writer.Close()
	request, err := http.NewRequest("POST", url, body)

	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Add("Authorization", API_KEY)

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	return content
}

func SendPostRequest(url string, param map[string]string) []byte {

	var r http.Request
	_ = r.ParseForm()
	for key, value := range param {
		r.Form.Add(key, value)
	}
	bodystr := strings.TrimSpace(r.Form.Encode())

	request, err := http.NewRequest("POST", url, strings.NewReader(bodystr))

	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("Authorization", API_KEY)
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	return content
}

func SendDownloadRequest(url string, param map[string]string, filepath string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil  {
		return err
	}
	defer out.Close()

	var r http.Request
	_ = r.ParseForm()
	for key, value := range param {
		r.Form.Add(key, value)
	}
	bodystr := strings.TrimSpace(r.Form.Encode())

	request, err := http.NewRequest("POST", url, strings.NewReader(bodystr))

	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("Authorization", API_KEY)
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil  {
		return err
	}

	return nil
}