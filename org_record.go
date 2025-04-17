package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	queries "github.com/adisuper94/orcidparser/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
var cacheMutex = &sync.RWMutex{}

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
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if len(gridCache) > 200_000 || len(ringgoldCache) > 200_000 || len(rorCache) > 200_000 {
		log.Println("flushing cache")
		gridCache = nil
		rorCache = nil
		ringgoldCache = nil
		InitCache()
	}
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
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
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
	orgName := pgtype.Text{String: org.Name, Valid: true}
	orgCountry := pgtype.Text{String: org.Address.Country, Valid: true}
	orgCity := pgtype.Text{String: org.Address.City, Valid: true}
	var orgRegion = pgtype.Text{String: "", Valid: false}
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
			params.GridID = pgtype.Text{String: id, Valid: true}
			insertParams.GridID = pgtype.Text{String: id, Valid: true}
		case "ROR":
			params.RorID = pgtype.Text{String: id, Valid: true}
			insertParams.RorID = pgtype.Text{String: id, Valid: true}
		case "FUNDREF":
			// params.FundrefID = pgtype.Text{String: id, Valid: true}
			// insertParams.FundrefID = pgtype.Text{String: id, Valid: true}
		case "LEI":
			params.LeiID = pgtype.Text{String: id, Valid: true}
			insertParams.LeiID = pgtype.Text{String: id, Valid: true}
		default:
			fmt.Println("Found a new type of org type: ", src, "id : ", id)
		}
	}
	orgRow, err := q.GetOrg(ctx, params)
	switch err {
	case pgx.ErrNoRows:
		orgRow, err = q.InsertOrg(ctx, insertParams)
		if err == nil {
			updateCache(orgRow, rid)
		} else {
			log.Fatalln(err, params, orgRow)
		}
	case nil:
		if org.DisambiguatedOrganization != nil {
			update := true
			updateParams := queries.UpdateOrgIdsParams{
				ID:     orgRow.ID,
				GridID: orgRow.GridID,
				RorID:  orgRow.RorID,
				// FundrefID: orgRow.FundrefID,
				LeiID: orgRow.LeiID,
			}
			src := org.DisambiguatedOrganization.Source
			id := org.DisambiguatedOrganization.Identifier
			switch src {
			case "RINGGOLD":
				update = false
			case "GRID":
				update = orgRow.GridID.String != id
				updateParams.GridID = pgtype.Text{String: id, Valid: true}
			case "ROR":
				update = orgRow.RorID.String != id
				updateParams.RorID = pgtype.Text{String: id, Valid: true}
			case "FUNDREF":
				// 	updateParams.FundrefID = pgtype.Text{String: id, Valid: true}
				return orgRow, err
			case "LEI":
				update = orgRow.LeiID.String != id
				updateParams.LeiID = pgtype.Text{String: id, Valid: true}
			default:
				log.Println("UpsertOrg: no clue what got updated")
				return orgRow, err
			}
			if update {
				orgRow, err = q.UpdateOrgIds(ctx, updateParams)
				if err == nil {
				} else {
					log.Fatalln(err, "\norg:", orgRow, "\nparams:", updateParams)
				}
			}
			updateCache(orgRow, rid)
		}
	default:
		log.Fatalln("wtf: unknown err: ", err)
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
