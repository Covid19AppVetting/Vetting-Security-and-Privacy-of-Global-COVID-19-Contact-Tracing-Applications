package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

const API_KEY = "401d30560609aa90e810f6818cb9c027823b5f03c1708d0270c1c5cd2b4307c8"

const BASE_PATH = "/Users/zachwang/Develop/assignments/MPhil/Contact-Tracing-Apps/Contact Tracing Apks"
const SAVING_PATH = "/Users/zachwang/Develop/assignments/MPhil/Contact-Tracing-Apps/reports/FlowDroid"

var quit chan int = make(chan int, 4)
var wg = sync.WaitGroup{}

type MyFile struct {
	path string
	filename string
}

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

	//for i := 0; i < jobsCount; i++ {
	//	fmt.Printf("index: %d,goroutine Num: %d\n", i, runtime.NumGoroutine())
	//}


	wg.Wait()
	close(quit)
	fmt.Println("all done")
}

func uploadApk(filepath string, filename string){

	defer wg.Done()

	fmt.Println("analyzing", filename)

	savingPath := SAVING_PATH + filepath[len(BASE_PATH):len(filepath)-3] + "xml"

	if _, err := os.Stat(savingPath); os.IsExist(err) {
		_ = os.Remove(savingPath)
	}
	savingParentDir := SAVING_PATH + filepath[len(BASE_PATH):len(filepath)-len(filename)]
	if _, err := os.Stat(savingParentDir); os.IsNotExist(err) {
		_ = os.MkdirAll(savingParentDir, 0777)
	}
	t1 := time.Now() // get current time
	cmd := exec.Command("java", "-jar",
		"/Users/zachwang/Develop/assignments/MPhil/FlowDroid/soot-infoflow-cmd-jar-with-dependencies.jar",
		"-a", filepath,
		"-p", "/Users/zachwang/Library/Android/sdk/platforms",
		"-s", "/Users/zachwang/Develop/assignments/MPhil/FlowDroid/soot-infoflow-android/SourcesAndSinks.txt",
		"-o", savingPath)
	err := cmd.Run()
	if err != nil{
		fmt.Println("error", err.Error())
	}
	elapsed := time.Since(t1)
	fmt.Println("done", filename, "time:", elapsed.Seconds())
}



