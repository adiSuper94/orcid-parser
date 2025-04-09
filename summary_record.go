package main

import "encoding/xml"

type Record struct {
	XMLName           xml.Name           `xml:"record"`
	Path              string             `xml:"path,attr"`
	OrcidIdentifier   *OrcidIdentifier   `xml:"orcid-identifier,omitempty"`
	Person            *Person            `xml:"person,omitempty"`
	ActivitiesSummary *ActivitiesSummary `xml:"activities-summary,omitempty"`
}

type ActivitiesSummary struct {
	XMLName          xml.Name     `xml:"activities-summary"`
	Path             string       `xml:"path,attr"`
	LastModifiedDate string       `xml:"last-modified-date,omitempty"`
	Educations       *Educations  `xml:"educations,omitempty"`
	Employments      *Employments `xml:"employments,omitempty"`
}

type Educations struct {
	XMLName          xml.Name           `xml:"educations"`
	Path             string             `xml:"path,attr"`
	LastModifiedDate string             `xml:"last-modified-date,omitempty"`
	AffiliationGroup []AffiliationGroup `xml:"affiliation-group,omitempty"`
}

type Employments struct {
	XMLName          xml.Name           `xml:"employments"`
	Path             string             `xml:"path,attr"`
	LastModifiedDate string             `xml:"last-modified-date,omitempty"`
	AffiliationGroup []AffiliationGroup `xml:"affiliation-group,omitempty"`
}

// Affiliation group contains either education or employment summaries
type AffiliationGroup struct {
	XMLName           xml.Name     `xml:"affiliation-group"`
	LastModifiedDate  string       `xml:"last-modified-date,omitempty"`
	EducationSummary  *Education   `xml:"education-summary,omitempty"`
	EmploymentSummary *Employment  `xml:"employment-summary,omitempty"`
}
