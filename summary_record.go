package main

import (
	"archive/tar"
	"context"
	"encoding/xml"
	"fmt"
	"log"
)

type RecordOption struct {
	Record *Record        `xml:"record,omitempty"`
	Error  *ErrorResponse `xml:"error,omitempty"`
}

// Structure for Error response
type ErrorResponse struct {
	ResponseCode     int    `xml:"response-code"`
	DeveloperMessage string `xml:"developer-message"`
	UserMessage      string `xml:"user-message"`
	ErrorCode        int    `xml:"error-code"`
	MoreInfo         string `xml:"more-info"`
}

type Record struct {
	Path              string             `xml:"path,attr"`
	OrcidIdentifier   OrcidIdentifier    `xml:"orcid-identifier,omitempty"`
	Person            *Person            `xml:"person,omitempty"`
	ActivitiesSummary *ActivitiesSummary `xml:"activities-summary,omitempty"`
}

type ActivitiesSummary struct {
	Path             string           `xml:"path,attr"`
	LastModifiedDate string           `xml:"last-modified-date,omitempty"`
	Educations       *ActivitySummary `xml:"educations,omitempty"`
	Employments      *ActivitySummary `xml:"employments,omitempty"`
}

type ActivitySummary struct {
	Path             string             `xml:"path,attr"`
	LastModifiedDate string             `xml:"last-modified-date,omitempty"`
	AffiliationGroup []AffiliationGroup `xml:"affiliation-group,omitempty"`
}

type AffiliationGroup struct {
	LastModifiedDate  string      `xml:"last-modified-date,omitempty"`
	EducationSummary  *Education  `xml:"education-summary,omitempty"`
	EmploymentSummary *Employment `xml:"employment-summary,omitempty"`
}

func ParseSummaryRecord(header *tar.Header, record *tar.Reader) Record {
	decoder := xml.NewDecoder(record)
	var summaryRecord Record
	err := decoder.Decode(&summaryRecord)
	if err != nil {
		log.Fatalln("Error Decoding", header.Name, ". Err: ", err)
	}
	return summaryRecord
}

func (r Record) String() string {
	return fmt.Sprintf("{OrcidIdentifier: %v, Person: %v, ActivitiesSummary: %v}", r.OrcidIdentifier, r.Person, r.ActivitiesSummary)
}

func (a ActivitiesSummary) String() string {
	return fmt.Sprintf("{Path: %s, Educations: %v, Employments: %v}", a.Path, a.Educations, a.Employments)
}

func (e ActivitySummary) String() string {
	str := fmt.Sprintf("{Path: %s, LastModifiedDate: %s", e.Path, e.LastModifiedDate)
	if e.AffiliationGroup == nil {
		return str + "}"
	}
	str += ", AffiliationGroup: ["
	for _, group := range e.AffiliationGroup {
		str += fmt.Sprintf("%v, ", group)
	}
	str += "]}"
	return str
}

func (a AffiliationGroup) String() string {
	str := fmt.Sprintf("{LastModifiedDate: %s", a.LastModifiedDate)
	if a.EducationSummary != nil {
		str += ", EducationSummary: " + fmt.Sprintf("%v", a.EducationSummary)
	}
	if a.EmploymentSummary != nil {
		str += ", EmploymentSummary: " + fmt.Sprintf("%v", a.EmploymentSummary)
	}
	return str + "}"
}

func (a ActivitiesSummary) Upsert(ctx context.Context) {
	if a.Employments != nil {
		for _, group := range a.Employments.AffiliationGroup {
			group.Upsert(ctx)
		}
	}
	if a.Educations != nil {
		for _, group := range a.Educations.AffiliationGroup {
			group.Upsert(ctx)
		}
	}
}

func (a AffiliationGroup) Upsert(ctx context.Context) {
	if a.EducationSummary != nil {
		a.EducationSummary.Upsert(ctx)
	}
	if a.EmploymentSummary != nil {
		a.EmploymentSummary.Upsert(ctx)
	}
}

func (r Record) Upsert(ctx context.Context) {
	orcidId := r.OrcidIdentifier.Path
	if r.Person != nil {
		r.Person.Upsert(orcidId, ctx)
	}
	if r.ActivitiesSummary != nil {
		r.ActivitiesSummary.Upsert(ctx)
	}
}
