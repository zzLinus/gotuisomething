package reddit

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/zzLinus/GoTUITODOList/extractor"
	"github.com/zzLinus/GoTUITODOList/fakeheaders"

	"fmt"
	"log"

	"github.com/zzLinus/GoTUITODOList/utils"
)

const (
	redditAPI = "https://v.redd.it/"
	res720    = "/DASH_720.mp4"
)

func init() {
	fmt.Println("init function get called")
	extractor.Register("reddit", New())
}

type redditExtractor struct{}

func (*redditExtractor) ExtractRowURL(rowURL string) (*extractor.Data, error) {
	var fileType = ""
	html, err := getHTMLPage(rowURL)
	if err != nil {
		log.Fatal("Failed to get html page")
	}

	videoName := utils.MatchOneOf(html, `<title>.*<\/title>`)[0]
	if utils.MatchOneOf(html, `meta property="og:video" content=.*HLSPlaylist`)[0] != "" {
		fileType = "mp4"
	}
	url := utils.MatchOneOf(html, `https://v.redd.it/.*/HLSPlaylist`)[0]
	if url == "" {
		panic("can't match anything")
	}
	//
	// for i := len(url) - 1; i >= 0; i-- {
	// 	if url[i] == '/' {
	// 		url = url[:i]
	// 	}
	// }
	// for i := len(url) - 1; i >= 0; i-- {
	// 	if url[i] == '/' {
	// 		url = url[i+1:]
	// 	}
	// }
	// for i := len(videoName) - 1; i >= 0; i-- {
	// 	if videoName[i] == '<' {
	// 		videoName = videoName[:i]
	// 	}
	// }
	// for i := len(videoName) - 1; i >= 0; i-- {
	// 	if videoName[i] == '>' {
	// 		url = videoName[i+1:]
	// 	}
	// }

	url = fmt.Sprintf("%s%s%s", redditAPI, url[18:31], res720)
	fmt.Println(url)

	return &extractor.Data{
		DownloadableURL: url,
		FileType:        videoName,
		VideoName:       fileType,
	}, nil
}

func New() extractor.Extractor {
	return &redditExtractor{}
}

func getHTMLPage(rowURL string) (string, error) {
	var (
		reTrytimes = 10
		resp       = &http.Response{}
	)
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DisableCompression:  true,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Minute,
	}

	req, err := http.NewRequest(http.MethodGet, rowURL, nil)
	if err != nil {
		log.Fatal("failed at create request")
		return "", err
	}

	for k, v := range fakeheaders.FakeHeaders {
		req.Header.Set(k, v)
	}

	req.Header.Set("Referer", "https://www.reddit.com/")
	req.Header.Set("Origin", "https://www.reddit.com")

	for ; reTrytimes > 0; reTrytimes-- {
		resp, err = client.Do(req)
		if err != nil && reTrytimes > 0 {
			log.Fatal("failed to recive a response")
			return "", err
		} else {
			break
		}
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return string(body), nil
}
