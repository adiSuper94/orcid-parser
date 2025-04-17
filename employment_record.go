package main

import (
	"archive/tar"
	"context"
	"encoding/xml"
	"fmt"
	"log"

	queries "github.com/adisuper94/orcidparser/generated"
	"github.com/jackc/pgx/v5"
)

type Employment struct {
	PutCode      int64  `xml:"put-code,attr"`
	Path         string `xml:"path,attr"`
	DisplayIndex string `xml:"display-index,attr"`
	Visibility   string `xml:"visibility,attr"`

	CreatedDate      string       `xml:"created-date"`
	LastModifiedDate string       `xml:"last-modified-date"`
	Source           Source       `xml:"source"`
	DepartmentName   string       `xml:"department-name"`
	RoleTitle        string       `xml:"role-title"`
	StartDate        *Date        `xml:"start-date"`
	EndDate          *Date        `xml:"end-date"`
	Organization     Organization `xml:"organization"`
	URL              string       `xml:"url"`
}

func ParseEmploymentRecord(header *tar.Header, record *tar.Reader) Employment {
	decoder := xml.NewDecoder(record)
	var empRecord Employment
	err := decoder.Decode(&empRecord)
	if err != nil {
		log.Fatalln("Error Decoding", header.Name, ". Err: ", err)
	}
	return empRecord
}

func (record Employment) Upsert(ctx context.Context) (*queries.Employment, error) {
	org, err := record.Organization.Upsert(ctx)
	if err != nil {
		log.Fatalln("Error while upserting org from `Upsert in employment` record:", record, "err: ", err)
	}
	emp, err := q.GetEmployment(ctx, record.PutCode)
	switch err {
	case nil:
		return &emp, err
	case pgx.ErrNoRows:
		insertParams := queries.InsertEmpoymentRecordParams{
			ID: record.PutCode, OrcidID: record.Source.SourceOrcid.Path,
			OrgID: org.ID, DeptName: emp.DeptName, RoleTitle: emp.RoleTitle}
		if record.StartDate != nil {
			tyme, err := record.StartDate.ToTime()
			if err != nil {
				// log.Println("could not get emp start date for empid:", record.PutCode, "orcidId:", record.Source.SourceOrcid.Path, "err:", err)
			}
			insertParams.StartDate = tyme
		}
		if record.EndDate != nil {
			tyme, err := record.EndDate.ToTime()
			if err != nil {
				// log.Println("could not get emp end date for empid:", record.PutCode, "orcidId:", record.Source.SourceOrcid.Path, "err:", err, *record.EndDate)
			}
			insertParams.EndDate = tyme
		}
		emp, err := q.InsertEmpoymentRecord(ctx, insertParams)
		return &emp, err
	default:
		fmt.Println("wtf!!, emp:", emp, "err: ", err)
		return nil, err
	}
}
