package tmplink2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

// do others that not defined in Driver interface
func mustParseURL(rawURL string) *url.URL {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		panic("Error parsing URL: " + err.Error())
	}
	return parsedURL
}

// set request header from map
func setHeader(req *http.Request, header map[string]string) {
	for k, v := range header {
		req.Header.Set(k, v)
	}
}

func (d *Tmplink2) tokenToken() (string, error) {
	captcha, err := d.tokenChallenge()
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Set("action", "token")
	params.Set("captcha", captcha)
	body := bytes.NewBufferString(params.Encode())

	// Create a POST request with the buffer content
	request, err := http.NewRequest("POST", "https://tmp-api.vx-cdn.com/api_v2/token", body)
	if err != nil {
		return "", err
	}

	setHeader(request, map[string]string{
		"Host":               "tmp-api.vx-cdn.com",
		"Connection":         "keep-alive",
		"sec-ch-ua":          `"Not_A Brand";v="8", "Chromium";v="120"`,
		"Accept":             "*/*",
		"DNT":                "1",
		"sec-ch-ua-mobile":   "?0",
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"sec-ch-ua-platform": `"macOS"`,
		"Origin":             "https://www.tmp.link",
		"Sec-Fetch-Site":     "cross-site",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
		"Referer":            "https://www.tmp.link/",
		"Accept-Language":    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Content-Type":       "application/x-www-form-urlencoded; charset=UTF-8",
	})

	response, err := d.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var tokenResponse map[string]interface{}
	_ = json.NewDecoder(response.Body).Decode(&tokenResponse)

	return tokenResponse["data"].(string), nil
}

func (d *Tmplink2) listRooms(id string) ([]TmpLinkSubMeetingRoomData, error) {
	if id == "" {
		id = "0"
	}

	url := "https://tmp-api.vx-cdn.com/api_v2/meetingroom"

	payload := "action=details&token=" + d.Token + "&mr_id=" + id

	request, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	setHeader(request, map[string]string{
		"Host":               "tmp-api.vx-cdn.com",
		"Connection":         "keep-alive",
		"sec-ch-ua":          `"Not_A Brand";v="8", "Chromium";v="120"`,
		"Accept":             "*/*",
		"DNT":                "1",
		"sec-ch-ua-mobile":   "?0",
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"sec-ch-ua-platform": `"macOS"`,
		"Origin":             "https://www.tmp.link",
		"Sec-Fetch-Site":     "cross-site",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
		"Referer":            "https://www.tmp.link/",
		"Accept-Language":    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Content-Type":       "application/x-www-form-urlencoded; charset=UTF-8",
	})

	response, err := d.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var mrResponse TmpLinkMeetingRoomResponse
	_ = json.NewDecoder(response.Body).Decode(&mrResponse)

	return mrResponse.Data.SubRooms, nil
}

func (d *Tmplink2) listFilesInRoom(mr string) ([]TmpLinkFileData, error) {
	if mr == "" {
		return nil, nil
	}

	payload := "action=file_list_page&token=" + d.Token + "&page=all&photo=0&mr_id=" + mr + "&sort_by=1&sort_type=1&search="

	url := "https://tmp-api.vx-cdn.com/api_v2/meetingroom"

	request, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	setHeader(request, map[string]string{
		"Host":               "tmp-api.vx-cdn.com",
		"Connection":         "keep-alive",
		"sec-ch-ua":          `"Not_A Brand";v="8", "Chromium";v="120"`,
		"Accept":             "*/*",
		"DNT":                "1",
		"sec-ch-ua-mobile":   "?0",
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"sec-ch-ua-platform": `"macOS"`,
		"Origin":             "https://www.tmp.link",
		"Sec-Fetch-Site":     "cross-site",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
		"Referer":            "https://www.tmp.link/",
		"Accept-Language":    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Content-Type":       "application/x-www-form-urlencoded; charset=UTF-8",
	})

	response, err := d.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var fileResponse TmpLinkFileResponse
	_ = json.NewDecoder(response.Body).Decode(&fileResponse)

	return fileResponse.Data, nil
}

func (d *Tmplink2) login() error {
	tmp, err := d.tokenChallenge()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("action", "login")
	params.Set("token", d.Token)
	params.Set("captcha", tmp)
	params.Set("email", d.Username)
	params.Set("password", d.Password)
	body := bytes.NewBufferString(params.Encode())

	// Create a POST request with the buffer content
	request, err := http.NewRequest("POST", "https://tmp-api.vx-cdn.com/api_v2/user", body)
	if err != nil {
		return err
	}

	setHeader(request, map[string]string{
		"Host":               "tmp-api.vx-cdn.com",
		"Connection":         "keep-alive",
		"sec-ch-ua":          `"Not_A Brand";v="8", "Chromium";v="120"`,
		"Accept":             "*/*",
		"DNT":                "1",
		"sec-ch-ua-mobile":   "?0",
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"sec-ch-ua-platform": `"macOS"`,
		"Origin":             "https://www.tmp.link",
		"Sec-Fetch-Site":     "cross-site",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
		"Referer":            "https://www.tmp.link/",
		"Accept-Language":    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Content-Type":       "application/x-www-form-urlencoded; charset=UTF-8",
	})

	response, err := d.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

func (d *Tmplink2) tokenChallenge() (string, error) {
	url := "https://tmp-api.vx-cdn.com/api_v2/token"

	payload := "action=challenge"

	request, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return "", err
	}

	setHeader(request, map[string]string{
		"Host":               "tmp-api.vx-cdn.com",
		"Connection":         "keep-alive",
		"sec-ch-ua":          `"Not_A Brand";v="8", "Chromium";v="120"`,
		"Accept":             "*/*",
		"DNT":                "1",
		"sec-ch-ua-mobile":   "?0",
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"sec-ch-ua-platform": `"macOS"`,
		"Origin":             "https://www.tmp.link",
		"Sec-Fetch-Site":     "cross-site",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
		"Referer":            "https://www.tmp.link/",
		"Accept-Language":    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Content-Type":       "application/x-www-form-urlencoded; charset=UTF-8",
	})

	response, err := d.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	//parse to map
	var tokenResponse map[string]interface{}
	_ = json.NewDecoder(response.Body).Decode(&tokenResponse)

	return tokenResponse["data"].(string), nil
}

func (d *Tmplink2) moveFileToMeetingRoom(ukey, mr string) error {
	payload := "action=move_to_dir&token=" + d.Token + "&ukey%5B%5D=" + ukey + "&mr_id=" + mr

	url := "https://tmp-api.vx-cdn.com/api_v2/meetingroom"

	request, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return err
	}

	setHeader(request, map[string]string{
		"Host":               "tmp-api.vx-cdn.com",
		"Connection":         "keep-alive",
		"sec-ch-ua":          `"Not_A Brand";v="8", "Chromium";v="120"`,
		"Accept":             "*/*",
		"DNT":                "1",
		"sec-ch-ua-mobile":   "?0",
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"sec-ch-ua-platform": `"macOS"`,
		"Origin":             "https://www.tmp.link",
		"Sec-Fetch-Site":     "cross-site",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
		"Referer":            "https://www.tmp.link/",
		"Accept-Language":    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Content-Type":       "application/x-www-form-urlencoded; charset=UTF-8",
	})

	response, err := d.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}
