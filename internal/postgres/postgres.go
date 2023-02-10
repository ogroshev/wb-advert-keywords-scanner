package postgres

import (
	"database/sql"
	"fmt"
	"strings"
)

type Category string

const (
	Strong   Category = "strong"
	Excluded Category = "excluded"
	Plus     Category = "plus"
	Keywords  Category = "keywords"
)

type AdvCompany struct {
	Id         uint64
	SellerID   uint64
	Name       string
	Cookie     string
	XUserID    string
	LastScanTs sql.NullTime
}

type Keyword struct {
	AdvertCompanyID uint64
	Keyword         string
	Category        Category
}

func GetActiveAdvCompanies(db *sql.DB) ([]AdvCompany, error) {
	q := `
		SELECT ac.id, ac.id_seller, ac.name, s.cpm_cookies, s.x_user_id, ak.update_dt AT TIME ZONE 'MSK' as statwords_last_update
		FROM advert_company ac
		JOIN sellers s ON ac.id_seller = s.id
		LEFT JOIN advert_keyword ak ON ac.id = ak.advert_company_id
			AND ak.id = (SELECT MAX(id)
						FROM advert_keyword ak2
						WHERE ak2.advert_company_id = ac.id)
		WHERE ac.turn_scan = TRUE;	
	`
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("processing query: %s. error: %s", q, err)
	}
	defer rows.Close()

	advCompanies := []AdvCompany{}
	for rows.Next() {
		ac := AdvCompany{}
		err := rows.Scan(&ac.Id, &ac.SellerID, &ac.Name, &ac.Cookie, &ac.XUserID, &ac.LastScanTs)
		if err != nil {
			return nil, fmt.Errorf("scan row: %s", err)
		}
		advCompanies = append(advCompanies, ac)
	}
	return advCompanies, nil
}

func SaveAdvertStatWords(db *sql.DB, adv_company_id uint64, statWords []string) error {
	values := make([]string, 0, len(statWords))
	for _, value := range statWords {
		values = append(values, fmt.Sprintf("(%d, '%s')", adv_company_id, value))
	}
	q := fmt.Sprintf("INSERT INTO advert_keyword (advert_company_id, keyword) VALUES %s", strings.Join(values, ","))

	_, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("processing query: %s. error: %s", q, err)
	}
	return nil
}

func UpsertAdvertStatWords(db *sql.DB, adv_company_id uint64, keywords []Keyword) error {
	values := make([]string, 0, len(keywords))
	for _, value := range keywords {
		values = append(values, fmt.Sprintf("(%d, '%s', '%s')", value.AdvertCompanyID, value.Keyword, value.Category))
	}
	q := fmt.Sprintf(`
		INSERT INTO advert_keyword (advert_company_id, keyword, category)
		VALUES %s 
		ON CONFLICT (advert_company_id, keyword) DO UPDATE SET 
		category = excluded.category,
		update_dt = now()
	`, strings.Join(values, ","))

	_, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("processing query: %s. error: %s", q, err)
	}
	return nil
}

func GetAdvertStatWords(db *sql.DB, adv_company_id uint64) ([]string, error) {
	q := `
		SELECT keyword
		FROM advert_keyword
		WHERE advert_company_id = $1
	`
	rows, err := db.Query(q, adv_company_id)
	if err != nil {
		return nil, fmt.Errorf("processing query: %s. error: %s", q, err)
	}
	defer rows.Close()

	statWords := []string{}
	for rows.Next() {
		var statWord string
		err := rows.Scan(&statWord)
		if err != nil {
			return nil, fmt.Errorf("scan row: %s", err)
		}
		statWords = append(statWords, statWord)
	}
	return statWords, nil
}
