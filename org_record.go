package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	queries "github.com/adisuper94/orcidparser/generated"
)

type Organization struct {
	Name    string `xml:"name"`
	Address struct {
		City    string `xml:"city"`
		Region  string `xml:"region,omitempty"`
		Country string `xml:"country"`
	} `xml:"address"`
	DisambiguatedOrganization *struct {
		Identifier string `xml:"disambiguated-organization-identifier"`
		Source     string `xml:"disambiguation-source"`
	} `xml:"disambiguated-organization,omitempty"`
}

func UpsertOrg(record Organization, ctx context.Context) (queries.Org, error) {
	q := GetQueries()
	orgName := sql.NullString{String: record.Name, Valid: true}
	orgCountry := sql.NullString{String: record.Address.Country, Valid: true}
	orgCity := sql.NullString{String: record.Address.City, Valid: true}
	var orgRegion = sql.NullString{String: "", Valid: false}
	if record.Address.Region != "" {
		orgRegion.String = record.Address.Region
		orgRegion.Valid = true
	}
	params := queries.GetOrgParams{Name: orgName, Country: orgCountry}
	insertParams := queries.InsertOrgParams{Name: orgName, Country: orgCountry, Region: orgRegion, City: orgCity}
	if record.DisambiguatedOrganization != nil {
		src := record.DisambiguatedOrganization.Source
		id := record.DisambiguatedOrganization.Identifier
		switch src {
		case "RINGGOLD":
		case "GRID":
			params.GridID = sql.NullString{String: id, Valid: true}
			insertParams.GridID = sql.NullString{String: id, Valid: true}
		case "ROR":
			params.RorID = sql.NullString{String: id, Valid: true}
			insertParams.RorID = sql.NullString{String: id, Valid: true}
		case "FUNDREF":
			params.FundrefID = sql.NullString{String: id, Valid: true}
			insertParams.FundrefID = sql.NullString{String: id, Valid: true}
		case "LEI":
			params.LeiID = sql.NullString{String: id, Valid: true}
			insertParams.LeiID = sql.NullString{String: id, Valid: true}
		default:
			fmt.Println("Found a new type of org type: ", src, "id : ", id)
		}
	}
	org, err := q.GetOrg(ctx, params)
	switch err {
	case sql.ErrNoRows:
		org, err = q.InsertOrg(ctx, insertParams)
	case nil:
		if record.DisambiguatedOrganization != nil {
			updateParams := queries.UpdateOrgIdsParams{
				ID:        org.ID,
				GridID:    org.GridID,
				RorID:     org.RorID,
				FundrefID: org.FundrefID,
				LeiID:     org.LeiID,
			}
			src := record.DisambiguatedOrganization.Source
			id := record.DisambiguatedOrganization.Identifier
			switch src {
			case "RINGGOLD":
				return org, err
			case "GRID":
				updateParams.GridID = sql.NullString{String: id, Valid: true}
			case "ROR":
				updateParams.RorID = sql.NullString{String: id, Valid: true}
			case "FUNDREF":
				updateParams.FundrefID = sql.NullString{String: id, Valid: true}
			case "LEI":
				updateParams.LeiID = sql.NullString{String: id, Valid: true}
			default:
				log.Println("UpsertOrg: no clue what got updated")
				return org, err
			}
			q.UpdateOrgIds(ctx, updateParams)
		}
	}
	return org, err
}
