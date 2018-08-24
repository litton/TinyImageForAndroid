package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	//"strings"
)

const (
	tinyApiHost = "https://api.tinify.com"
)

var apiKey = flag.String("apiKey", "0y6rVN4iMdUTFjIa85P4QxgG9kZtIOIs", "api key for accesss tiny")
var img = flag.String("img", "", "compressed image path")
var outdir = flag.String("out", "./", "the dir of download compressed image")

type TinyHandler struct {
	ApiKey      string
	httpclient  *http.Client
	authortoken string
	outImgDir   string
}

func (tinyhandler *TinyHandler) InitHandler(apikey string, outImgDir string) {
	tinyhandler.ApiKey = apikey
	tinyhandler.httpclient = http.DefaultClient
	tinyhandler.outImgDir = outImgDir
	tinyhandler.authortoken = tinyhandler.getAuthorCode(*apiKey)

}

func (tinyhandler *TinyHandler) getAuthorCode(apiKey string) string {
	apiStr := "api:" + apiKey
	encoded := base64.StdEncoding.EncodeToString([]byte(apiStr))
	return "Basic " + encoded
}

func (tinyhandler *TinyHandler) UploadFile(imagefilepath string) (string, error) {
	apiuri := tinyApiHost + "/shrink"
	imgBytes, err := ioutil.ReadFile(imagefilepath)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", apiuri, bytes.NewReader(imgBytes))
	if err != nil {
		return "", err
	}

	request.Header.Set("Content-Type", "multipart/form-data")
	request.Header.Set("Authorization", tinyhandler.authortoken)
	response, err := tinyhandler.httpclient.Do(request)
	if err != nil {
		return "", err
	}

	//fmt.Printf("response=%v", response)
	defer response.Body.Close()
	//body, err := ioutil.ReadAll(response.Body)
	imgUrl := response.Header.Get("Location")
	//log.Println("img url", imgUrl)
	//resBytes, err := ioutil.ReadAll(response.Body)
	//log.Println("res=" + string(resBytes))

	return imgUrl, nil
}

func (tinyhandler *TinyHandler) DownloadImg(imgUrl string) (string, error) {

	request, err := http.NewRequest("GET", imgUrl, nil)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	request.Header.Set("Authorization", tinyhandler.authortoken)
	response, err := tinyhandler.httpclient.Do(request)
	if err != nil {
		return "", err
	}
	resBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(tinyhandler.outImgDir, resBytes, 0644)
	defer response.Body.Close()

	return tinyhandler.outImgDir, nil

}

//ch := make(chan string)

func main() {
	flag.Parse()
	if len(*img) == 0 {
		log.Fatal("params img cannot is nil")
		return
	}

	if len(*outdir) == 0 {
		log.Fatal("you should input the output file path")
		return
	}
	ch := make(chan string)
	imgpath := *img
	imgName := filepath.Base(imgpath)
	log.Println("imgName=" + imgName)
	log.Println("outputfilename=" + filepath.Base(*outdir))
	tinyHandler := new(TinyHandler)
	tinyHandler.InitHandler(*apiKey, *outdir)
	//fmt.Println("vim-go")
	go func(tinyHandler *TinyHandler) {
		imgUrl, _ := tinyHandler.UploadFile(*img)
		log.Println(imgUrl)
		newfilename, _ := tinyHandler.DownloadImg(imgUrl)
		ch <- newfilename
	}(tinyHandler)

	filename := <-ch
	fmt.Println("generate new file " + filename)
}
