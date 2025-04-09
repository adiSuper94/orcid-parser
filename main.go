package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/adisuper94/orcidparser/generated"
	_ "modernc.org/sqlite"
)

func tarLs(filePath string, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		log.Fatal(err)
	}
	defer gzr.Close()

	ctx := context.Background()
	if err != nil {
		log.Println(err)
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
		buildIndex(header, tr, ctx)
	}
}

func buildIndex(header *tar.Header, record *tar.Reader, ctx context.Context) {
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
		q.InsertDir(ctx, queries.InsertDirParams{Name: sql.NullString{String: dirName, Valid: true}, ArchiveFileID: archiveFile.ID})
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
			employment, err := UpsertEmploymentRecord(employmentXML, ctx)
			if err != nil {
				log.Fatalln("Error while upserting employment", employmentXML, "Error: ", err, "inserted employement: ", employment)
			}
		case "educations":
			educationXML := ParseEducationRecord(header, record)
			if educationXML.Organization != nil {
				UpsertOrg(*educationXML.Organization, ctx)
			}
		case "membership", "peer-reviews", "works", "distinctions", "fundings", "qualifications", "services", "invited-positions", "research-resources":
		default:
			fmt.Println("unknow section:", section, "orcid:", orcidId, "file name:", fileName)
		}
	}
}

func main() {
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	InitCache()
	go tarLs("/home/adisuper/Downloads/ORCID_2024_10_activities_0.tar.gz", &wg)
}
