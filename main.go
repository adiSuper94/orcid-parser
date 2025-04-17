package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adisuper94/orcidparser/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

func tarLs(filePath string, dir int, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		log.Fatalln(err)
	}
	defer gzr.Close()

	ctx := context.Background()
	if err != nil {
		log.Fatalln(err)
	}
	if err != nil {
		log.Fatalln(err)
	}
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err != nil {
			break // EOF
		}
		buildSummaryIndex(header, tr, dir, ctx)
	}
}

func _buildIndex(header *tar.Header, record *tar.Reader, ctx context.Context) {
	steps := strings.Split(header.Name, "/")
	if len(steps) < 2 {
		return
	}
	q := GetQueries()
	archiveFileName := steps[0]
	dirName := steps[1]
	if len(steps) == 2 {
		log.Println("Started Parsing archive file: ", archiveFileName)
		q.InsertArchive(ctx, archiveFileName)
	} else if len(steps) == 3 {
		archiveFile, err := q.GetArchiveFile(ctx, archiveFileName)
		if err != nil {
			log.Fatalln("ERR: archive file not found", err)
		}
		log.Println("Started archive file:", archiveFileName, "dir: ", dirName)
		q.InsertDir(ctx, queries.InsertDirParams{Name: pgtype.Text{String: dirName, Valid: true}, ArchiveFileID: archiveFile.ID})
	} else if len(steps) == 5 {
		if header.Typeflag != tar.TypeReg { // only regular files
			return
		}
		orcidId := steps[2]
		section := steps[3]
		fileName := steps[4]
		switch section {
		case "employments":
			employmentXML := ParseEmploymentRecord(header, record)
			employment, err := employmentXML.Upsert(ctx)
			if err != nil {
				log.Fatalln("Error while upserting employment", employmentXML, "Error: ", err, "inserted employement: ", employment)
			}
		case "educations":
			educationXML := ParseEducationRecord(header, record)
			if educationXML.Organization != nil {
				educationXML.Organization.Upsert(ctx)
			}
		case "membership", "peer-reviews", "works", "distinctions", "fundings", "qualifications", "services", "invited-positions", "research-resources":
		default:
			fmt.Println("unknow section:", section, "orcid:", orcidId, "file name:", fileName)
		}
	}
}

func buildSummaryIndex(header *tar.Header, record *tar.Reader, dir int, ctx context.Context) {
	steps := strings.Split(header.Name, "/")
	if len(steps) != 3 {
		return
	}
	dirName := steps[1]
	secondChar, err := strconv.Atoi(dirName[1:2])
	if dir != secondChar && err == nil {
		return
	}
	firstChar, err := strconv.Atoi(dirName[0:1])
	if firstChar < 5 && err == nil {
		return
	}
	if header.Typeflag != tar.TypeReg { // only regular files
		dif := time.Now().Unix() - tyme.Unix()
		min := dif / 60
		sec := dif % 60
		tyme = time.Now()
		fmt.Println(idx, "\t", min, "min", sec, "sec\t\t", header.Name)
		idx++
		return
	}
	summary := ParseSummaryRecord(header, record)
	summary.Upsert(ctx)
}

var tyme time.Time
var idx int32

func main() {
	tyme = time.Now()
	idx = 1
	var wg sync.WaitGroup
	defer wg.Wait()
	InitCache()
	for i := range 10 {
		wg.Add(1)
		go tarLs("/home/adisuper/Downloads/ORCID_2024_10_summaries.tar.gz", i, &wg)
	}
}
