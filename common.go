package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"time"

	queries "github.com/adisuper94/orcidparser/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrcidIdentifier struct {
	XMLName xml.Name `xml:"orcid-identifier"`
	URI     string   `xml:"uri"`
	Path    string   `xml:"path"`
	Host    string   `xml:"host"`
}

type Person struct {
	XMLName          xml.Name    `xml:"person"`
	Path             string      `xml:"path,attr"`
	LastModifiedDate string      `xml:"last-modified-date,omitempty"`
	Name             *PersonName `xml:"name,omitempty"`
	Emails           *Emails     `xml:"emails,omitempty"`
}

type Emails struct {
	XMLName          xml.Name `xml:"emails"`
	Path             string   `xml:"path,attr"`
	LastModifiedDate string   `xml:"last-modified-date,omitempty"`
	Email            []Email  `xml:"email,omitempty"`
}

type Email struct {
	XMLName          xml.Name `xml:"email"`
	Visibility       string   `xml:"visibility,attr"`
	Verified         bool     `xml:"verified,attr"`
	Primary          bool     `xml:"primary,attr"`
	CreatedDate      string   `xml:"created-date"`
	LastModifiedDate string   `xml:"last-modified-date"`
	Source           Source   `xml:"source"`
	Email            string   `xml:"email"`
}

type PersonName struct {
	XMLName          xml.Name `xml:"name"`
	Visibility       string   `xml:"visibility,attr"`
	Path             string   `xml:"path,attr"`
	CreatedDate      string   `xml:"created-date"`
	LastModifiedDate string   `xml:"last-modified-date"`
	GivenNames       string   `xml:"given-names"`
	FamilyName       string   `xml:"family-name"`
}

type Source struct {
	SourceOrcid struct {
		URI  string `xml:"uri"`
		Path string `xml:"path"`
		Host string `xml:"host"`
	} `xml:"source-orcid"`
	SourceName string `xml:"source-name"`
}

type Date struct {
	Year  *string `xml:"year"`
	Month *string `xml:"month"`
	Day   *string `xml:"day"`
}

func (d Date) ToMillis() (int64, error) {
	// Year is required
	if d.Year == nil {
		return 0, errors.New("year is required")
	}

	year, err := strconv.Atoi(*d.Year)
	if err != nil {
		return 0, errors.New("invalid year")
	}

	// Month defaults to January
	month := time.January
	if d.Month != nil {
		monthInt, err := strconv.Atoi(*d.Month)
		if err != nil || monthInt < 1 || monthInt > 12 {
			return 0, errors.New("invalid month")
		}
		month = time.Month(monthInt)
	}

	// Day defaults to 1
	day := 1
	if d.Day != nil {
		dayInt, err := strconv.Atoi(*d.Day)
		if err != nil || dayInt < 1 || dayInt > 31 {
			return 0, errors.New("invalid day")
		}
		day = dayInt
	}

	// Construct time
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return t.UnixMilli(), nil
}

func (d Date) ToTime() (*time.Time, error) {
	if d.Year == nil {
		return nil, errors.New("year is required")
	}
	if d.Month == nil {
		month := "01"
		d.Month = &month
	}
	dateString := fmt.Sprintf("%s-%s-%s", *d.Year, *d.Month, "01")
	layout := "2006-01-02"
	tyme, err := time.Parse(layout, dateString)
	if err != nil {
		return nil, err
	}
	return &tyme, err
}

func (p Person) Upsert(orcidId string, ctx context.Context) (queries.Person, error) {
	q := GetQueries()
	givenName := pgtype.Text{String: "", Valid: false}
	familyName := pgtype.Text{String: "", Valid: false}
	if p.Name == nil {
		return queries.Person{}, errors.New("person name is nil")
	}
	if p.Name.GivenNames != "" {
		givenName = pgtype.Text{String: p.Name.GivenNames, Valid: true}
	}
	if p.Name.FamilyName != "" {
		familyName = pgtype.Text{String: p.Name.FamilyName, Valid: true}
	}
	insertPersonParams := queries.InsertPersonParams{OrcidID: orcidId, GivenName: givenName, FamilyName: familyName}
	return q.InsertPerson(ctx, insertPersonParams)
}
