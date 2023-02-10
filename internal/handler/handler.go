package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"gitlab.com/wb-dynamics/wb-advert-keywords-scanner/internal/postgres"
	"gitlab.com/wb-dynamics/wb-advert-keywords-scanner/internal/wbrequest"
)

const (
	kIntervalBetweenCompaniesMillisec = 1000
)

func DoWork(db *sql.DB, scanIntervalSec int) {
	advCompanies, err := postgres.GetActiveAdvCompanies(db)
	if err != nil {
		log.Fatalf("could not get tasks: %s", err)
	}
	printReceivedAdvCompanyInfo(advCompanies)

	for _, ac := range advCompanies {
		if isItTimeToScan(ac.Id, ac.LastScanTs, scanIntervalSec) {
			log.Debugf("Scan stat-words for advert company: %d %s", ac.Id, ac.Name)
			words, err := wbrequest.GetStatWords(ac.Id, ac.Cookie, ac.XUserID)
			if err != nil {
				log.Printf("Could not get stat-words for advert company: %d. Skip...", ac.Id)
				continue
			}
			if !checkUniqueKeywords(*words) {
				log.Errorf("Words are not unique!")
				json, _ := json.Marshal(*words)
				log.Errorf("Words: %s", string(json))
				log.Fatalf("Fatal error. Keywords are not unique for advert company: %d", ac.Id)
			}
			statWords := wBWordsToDbKeywords(ac.Id, *words)

			if len(statWords) != 0 {
				err = postgres.UpsertAdvertStatWords(db, ac.Id, statWords)
				if err != nil {
					log.Fatalf("Could not save stat-words to database. Error: %s", err)
				}
				log.Debugf("Saved stat-words for advert company: %d", ac.Id)
			} else {
				log.Debugf("No stat-words for advert company: %d. Skip...", ac.Id)
			}
			log.Debugf("Sleep for %d msec between companies", kIntervalBetweenCompaniesMillisec)
			time.Sleep(time.Duration(kIntervalBetweenCompaniesMillisec) * time.Millisecond)
		}
	}
}

func isItTimeToScan(advCompanyID uint64, lastScanTs sql.NullTime, scanIntervalSec int) bool {
	if !lastScanTs.Valid {
		return true
	}

	last := lastScanTs.Time
	nextScanTs := last.Add(time.Second * time.Duration(scanIntervalSec))
	now := time.Now()
	log.Tracef("comapany id: %d, last scan: %v ", advCompanyID, last)
	log.Tracef("comapany id: %d, next scan: %v (now: %v)", advCompanyID, nextScanTs, now)

	return nextScanTs.Before(now)
}

func printReceivedAdvCompanyInfo(advCompanies []postgres.AdvCompany) {
	ids := make([]string, len(advCompanies))
	for i, ac := range advCompanies {
		ids[i] = fmt.Sprintf("%v", ac.Id)
	}
	log.Tracef("Selected %d advert companies for scan: %v", len(advCompanies), strings.Join(ids, ","))
}

func wBWordsToDbKeywords(advCompanyID uint64, words wbrequest.Words) (keyword []postgres.Keyword) {
	lengh := len(words.Strong) + len(words.Excluded) + len(words.Pluse) + len(words.Keywords)
	keyword = make([]postgres.Keyword, 0, lengh)
	for _, w := range words.Strong {
		keyword = append(keyword, postgres.Keyword{AdvertCompanyID: advCompanyID, Keyword: w, Category: postgres.Strong})
	}
	for _, w := range words.Excluded {
		keyword = append(keyword, postgres.Keyword{AdvertCompanyID: advCompanyID, Keyword: w, Category: postgres.Excluded})
	}
	for _, w := range words.Pluse {
		keyword = append(keyword, postgres.Keyword{AdvertCompanyID: advCompanyID, Keyword: w, Category: postgres.Plus})
	}
	for _, w := range words.Keywords {
		keyword = append(keyword, postgres.Keyword{AdvertCompanyID: advCompanyID, Keyword: w.Keyword, Category: postgres.Keywords})
	}
	return keyword
}

func checkUniqueKeywords(words wbrequest.Words) bool {
	uniqueKeywords := make(map[string]struct{})
	for _, k := range words.Strong {
		uniqueKeywords[k] = struct{}{}
	}
	for _, k := range words.Excluded {
		uniqueKeywords[k] = struct{}{}
	}
	for _, k := range words.Pluse {
		uniqueKeywords[k] = struct{}{}
	}
	for _, k := range words.Keywords {
		uniqueKeywords[k.Keyword] = struct{}{}
	}
	return len(uniqueKeywords) == len(words.Strong) + len(words.Excluded) + len(words.Pluse) + len(words.Keywords)
}