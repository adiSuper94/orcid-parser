package main

import (
	"archive/tar"
	"encoding/xml"
	"log"
)

type Education struct {
	PutCode      string `xml:"put-code,attr"`
	Visibility   string `xml:"visibility,attr"`
	DisplayIndex int    `xml:"display-index,attr"`
	Path         string `xml:"path,attr"`

	CreatedDate      string        `xml:"created-date"`
	LastModifiedDate string        `xml:"last-modified-date"`
	URL              string        `xml:"url,omitempty"`
	Source           *Source       `xml:"source"`
	DepartmentName   string        `xml:"department-name"`
	RoleTitle        string        `xml:"role-title"`
	StartDate        Date          `xml:"start-date"`
	EndDate          Date          `xml:"end-date"`
	Organization     *Organization `xml:"organization"`
}

func ParseEducationRecord(header *tar.Header, record *tar.Reader) Education {
	decoder := xml.NewDecoder(record)
	var eduRecord Education
	err := decoder.Decode(&eduRecord)
	if err != nil {
		log.Fatalln("Error Decoding", header.Name, ". Err: ", err)
	}
	return eduRecord
}
