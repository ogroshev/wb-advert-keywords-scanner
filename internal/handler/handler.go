package handler

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	log "github.com/sirupsen/logrus"
	
	"gitlab.com/wb-dynamics/wb-advert-keywords-scanner/internal/postgres"
	"gitlab.com/wb-dynamics/wb-advert-keywords-scanner/internal/wbrequest"
)

const (
	kIntervalBetweenCompaniesSec = 3
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
			keywords, err := wbrequest.GetStatWords(ac.Id, ac.Cookie, ac.XUserID)
			if err != nil {
				log.Printf("Could not get stat-words for advert company: %d. Skip...", ac.Id)
				continue
			}
			err = postgres.SaveAdvertStatWords(db, ac.Id, keywords)
			if err != nil {
				log.Fatalf("Could not save stat-words to database. Error: %s", err)
			}
			log.Debugf("Saved stat-words for advert company: %d", ac.Id)
			log.Tracef("Sleep for %d seconds", kIntervalBetweenCompaniesSec)
			time.Sleep(time.Duration(kIntervalBetweenCompaniesSec) * time.Second)
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