package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const API_KEY = "8aabd97fc1fa10b5f60ad710145a58a3d6a32f1b3d5f7bc04a3354c56fb6bd5f"
const BASE_URL = "http://localhost:8000/"


var logSensitive = map[string]bool{}
var crypto  = map[string]bool{}
var cleartextStorage  = map[string]bool{}
var sql  = map[string]bool{}
var insecureRandomValue  = map[string]bool{}
var allowBackup  = map[string]bool{}
var incorrectDefaultPermissions  = map[string]bool{}
var exposureSensitive  = map[string]bool{}
var cleartextTraffic  = map[string]bool{}
var insecureWeb  = map[string]bool{}
var inproperCertificate  = map[string]bool{}
var weaknessesInMobileApp  = map[string]bool{}
var launchmode  = map[string]bool{}

var md5s map[string]string = map[string]string{}

// for statistics
func main() {

	getAllApps()
}

type ScanOutput struct {
	Content []map[string]interface{}
	Count int
	Error string
}

type ReportObject struct {
	Version string
	AppName string `json:"app_name"`
	ExportedActivities interface{} `json:"exported_activities"`
	BrowsableActivities interface{} `json:"browsable_activities"`
	ManifestAnalysis []map[string]interface{} `json:"manifest_analysis"`
	CodeAnalysis map[string]map[string]interface{} `json:"code_analysis"`
}

func getAllApps(){
	fmt.Println("start")
	contentByte := SendGetRequest(BASE_URL + "api/v1/scans?page=1&page_size=50")
	var objmap ScanOutput //map[string]interface{}
	err := json.Unmarshal(contentByte, &objmap)
	if err != nil || objmap.Error != ""{
		fmt.Println(objmap.Error)
		return
	}


	var arrayObj = objmap.Content

	for _, value := range arrayObj {
		md5s[fmt.Sprint(value["MD5"])] = fmt.Sprint(value["FILE_NAME"])
	}

	fmt.Println("fetchAll done")

	analyseAll()

	fmt.Println("end")

	println("log", len(logSensitive))
	for md5, _ := range logSensitive {
		println("--", md5s[md5])
	}
	println("crypto", len(crypto))
	for md5, _ := range crypto {
		println("--", md5s[md5])
	}
	println("cleartextStorage", len(cleartextStorage))
	for md5, _ := range cleartextStorage {
		println("--", md5s[md5])
	}
	println("sql", len(sql))
	for md5, _ := range sql {
		println("--", md5s[md5])
	}
	println("insecureRandom", len(insecureRandomValue))
	for md5, _ := range insecureRandomValue {
		println("--", md5s[md5])
	}
	println("backup", len(allowBackup))
	for md5, _ := range allowBackup {
		println("--", md5s[md5])
	}
	println("incorrectPermissions", len(incorrectDefaultPermissions))
	for md5, _ := range incorrectDefaultPermissions {
		println("--", md5s[md5])
	}
	println("exposureSensitiveInformation", len(exposureSensitive))
	for md5, _ := range exposureSensitive {
		println("--", md5s[md5])
	}
	println("non-standard Launch Mode", len(launchmode))
	for md5, _ := range launchmode {
		println("--", md5s[md5])
	}
	println("cleartext Traffic", len(cleartextTraffic))
	for md5, _ := range cleartextTraffic {
		println("--", md5s[md5])
	}
	println("insecure Webview", len(insecureWeb))
	for md5, _ := range insecureWeb {
		println("--", md5s[md5])
	}
	println("improper Certificate", len(inproperCertificate))
	for md5, _ := range inproperCertificate {
		println("--", md5s[md5])
	}
	println("weakness in MobileApplication", len(weaknessesInMobileApp))
	for md5, _ := range weaknessesInMobileApp {
		println("--", md5s[md5])
	}


}

func analyseAll()  {
	for md5, _ := range md5s {
		contentByte := SendPostRequest(BASE_URL + "api/v1/report_json", map[string]string{"hash":md5})
		var objmap ReportObject //map[string]interface{}
		err := json.Unmarshal(contentByte, &objmap)
		if err != nil {
			fmt.Println("error", md5)
			return
		}

		//md5s[md5] = objmap.AppName

		for _, analysis := range objmap.ManifestAnalysis {

			title := fmt.Sprint(analysis["title"])
			if strings.Contains(title, "Application Data can be Backed up"){
				allowBackup[md5] = true
			}else if strings.Contains(title, "Clear text traffic"){
				cleartextTraffic[md5] = true
			}else if strings.Contains(title, "Launch Mode of Activity") && strings.Contains(title, "is not standard") {
				launchmode[md5] = true
			}
		}

		for key, value := range objmap.CodeAnalysis {
			if strings.Contains(key, "The App uses an insecure Random Number Generator") {
				insecureRandomValue[md5] = true
			}else if strings.Contains(key, "The App logs information") {
				logSensitive[md5] = true
			}else if strings.Contains(key, "App uses SQLite Database and execute raw SQL query") {
				sql[md5] = true
			}else if strings.Contains(strings.ToLower(key), strings.ToLower("Insecure WebView Implementation")){
				insecureWeb[md5] = true
			}else {
				cwe := fmt.Sprint(value["cwe"])
				if strings.Contains(cwe, "Weaknesses in Mobile Applications") {
					weaknessesInMobileApp[md5] = true
				}else if strings.Contains(cwe, "Improper Certificate Validation") {
					inproperCertificate[md5] = true
				}else if strings.Contains(cwe, "Exposure of Sensitive Information") {
					exposureSensitive[md5] = true
				}else if strings.Contains(cwe, "Incorrect Default Permissions") {
					incorrectDefaultPermissions[md5] = true
				}else if strings.Contains(cwe, "Cleartext Storage of Sensitive Information") {
					cleartextStorage[md5] = true
				}else if strings.Contains(cwe, "Use of a Broken or Risky Cryptographic Algorithm") {
					crypto[md5] = true
				}
			}
		}
	}
}

func SendGetRequest(url string) []byte {

	request, err := http.NewRequest("GET", url, nil)

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
