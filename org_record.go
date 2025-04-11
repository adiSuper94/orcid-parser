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

var gridCache map[string]*queries.Org
var rorCache map[string]*queries.Org
var ringgoldCache map[string]*queries.Org

func InitCache() {
	if gridCache == nil {
		gridCache = make(map[string]*queries.Org)
	}
	if rorCache == nil {
		rorCache = make(map[string]*queries.Org)
	}
	if ringgoldCache == nil {
		ringgoldCache = make(map[string]*queries.Org)
	}
}

func updateCache(org queries.Org, rid string) {
	if org.GridID.Valid {
		gridCache[org.GridID.String] = &org
	}
	if org.RorID.Valid {
		rorCache[org.RorID.String] = &org
	}
	if rid != "" {
		ringgoldCache[rid] = &org
	}
}

func cacheOut(src string, id string) (*queries.Org, bool) {
	switch src {
	case "GRID":
		if gridCache[id] != nil {
			return gridCache[id], true
		}
	case "ROR":
		if rorCache[id] != nil {
			return rorCache[id], true
		}
	case "RINGGOLD":
		if ringgoldCache[id] != nil {
			return ringgoldCache[id], true
		}
	}
	return nil, false
}

func (org Organization) Upsert(ctx context.Context) (queries.Org, error) {
	q := GetQueries()
	orgName := sql.NullString{String: org.Name, Valid: true}
	orgCountry := sql.NullString{String: org.Address.Country, Valid: true}
	orgCity := sql.NullString{String: org.Address.City, Valid: true}
	var orgRegion = sql.NullString{String: "", Valid: false}
	if org.Address.Region != "" {
		orgRegion.String = org.Address.Region
		orgRegion.Valid = true
	}
	params := queries.GetOrgParams{Name: orgName, Country: orgCountry}
	insertParams := queries.InsertOrgParams{Name: orgName, Country: orgCountry, Region: orgRegion, City: orgCity}
	rid := ""
	if org.DisambiguatedOrganization != nil {
		src := org.DisambiguatedOrganization.Source
		id := org.DisambiguatedOrganization.Identifier
		cachedOrg, ok := cacheOut(src, id)
		if ok {
			return *cachedOrg, nil
		}
		switch src {
		case "RINGGOLD":
			rid = id
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
	orgRow, err := q.GetOrg(ctx, params)
	switch err {
	case sql.ErrNoRows:
		orgRow, err = q.InsertOrg(ctx, insertParams)
		if err == nil {
			updateCache(orgRow, rid)
		}
	case nil:
		if org.DisambiguatedOrganization != nil {
			var rid = ""
			updateParams := queries.UpdateOrgIdsParams{
				ID:        orgRow.ID,
				GridID:    orgRow.GridID,
				RorID:     orgRow.RorID,
				FundrefID: orgRow.FundrefID,
				LeiID:     orgRow.LeiID,
			}
			src := org.DisambiguatedOrganization.Source
			id := org.DisambiguatedOrganization.Identifier
			switch src {
			case "RINGGOLD":
				rid = id
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
				return orgRow, err
			}
			if rid == "" {
				orgRow, err = q.UpdateOrgIds(ctx, updateParams)
			}
			if err == nil {
				updateCache(orgRow, rid)
			}
		}
	}
	return orgRow, err
}
func (o Organization) String() string {
	doStr := "nil"
	if o.DisambiguatedOrganization != nil {
		doStr = fmt.Sprintf("{Identifier: %s, Source: %s}", o.DisambiguatedOrganization.Identifier, o.DisambiguatedOrganization.Source)
	}
	return fmt.Sprintf("{Name: %s, Address: {City: %s, Region: %s, Country: %s}, DisambiguatedOrganization: %s", o.Name, o.Address.City, o.Address.Region, o.Address.Country, doStr)
}
