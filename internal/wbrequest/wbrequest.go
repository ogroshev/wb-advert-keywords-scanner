package wbrequest

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	wbr "gitlab.com/wb-dynamics/wb-go-libs/wbrequest"
)

const (
	kGetStatWords  = "https://cmp.wildberries.ru/backend/api/v2/search/%d/stat-words"
	kRefererHeader = "https://cmp.wildberries.ru/campaigns/list/active/edit/search/%d"
)

type keywordsResponse struct {
	Words struct {
		Phrase   []string `json:"phrase"`
		Strong   []string `json:"strong"`
		Excluded []string `json:"excluded"`
		Pluse    []string `json:"pluse"`
		Keywords []struct {
			Keyword string `json:"keyword"`
			Count   int    `json:"count"`
		} `json:"keywords"`
		Fixed bool `json:"fixed"`
	} `json:"words"`
	Stat []struct {
		AdvertId     int       `json:"advertId"`
		Keyword      string    `json:"keyword"`
		AdvertName   string    `json:"advertName"`
		CampaignName string    `json:"campaignName"`
		Begin        time.Time `json:"begin"`
		End          time.Time `json:"end"`
		Views        int       `json:"views"`
		Clicks       int       `json:"clicks"`
		Frq          float32   `json:"frq"`
		Ctr          float32   `json:"ctr"`
		Cpc          float32   `json:"cpc"`
		Duration     int       `json:"duration"`
		Sum          float32   `json:"sum"`
	} `json:"stat"`
	Total uint32 `json:"total"`
}

func GetStatWords(advCompanyID uint64, rawCookies string, xUserID string) (keywords []string, err error) {
	url := fmt.Sprintf(kGetStatWords, advCompanyID)
	headers := map[string]string{
		"Cookie":    rawCookies,
		"X-User-Id": xUserID,
		"Referer":   fmt.Sprintf(kRefererHeader, advCompanyID),
	}

	body, status_code, err := wbr.SendWithRetries("GET", url, headers)
	if err != nil {
		log.Errorf("http request 'stat-words' error: %v", err)
		return nil, err
	}
	log.Debugf("http request 'stat-words' status: %v", status_code)

	var keywordsResp keywordsResponse
	err = json.Unmarshal(body, &keywordsResp)
	if err != nil {
		log.Errorf("Could not parse json: %s err: %v", body, err)
		return nil, err
	}

	for _, kw := range keywordsResp.Words.Keywords {
		keywords = append(keywords, kw.Keyword)
	}
	return keywords, nil
}
