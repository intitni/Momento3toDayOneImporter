package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

const (
	momentoTimeStamp int64 = 476030880  // 02/01/2016 @ 2:48pm (UTC)
	unixTimeStamp    int64 = 1454338080 // 02/01/2016 @ 2:48pm (UTC)
	mediaPostfix           = "_original.jpg"
	journalPath            = "~/Library/Group Containers/5U8NS4GX82.dayoneapp2/Data/Auto Import/Default Journal.dayone"
)

var moments []*Moment
var photos []*Media
var filePath string
var randomUUIDPrefix string
var fileList []string

type Moment struct {
	ID          int64
	CreatedTime time.Time
	Note        string
	Media       []*Media
}

type Media struct {
	ID            int64
	MomentID      int64
	Identifier    string
	NewInMomento3 bool
}

func main() {
	currentPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	flag.StringVar(&filePath, "p", currentPath, "path of backup")
	flag.Parse()

	fmt.Println("Grabbing data from database....")

	grabData()

	fmt.Println(strconv.Itoa(len(moments)) + " moments and " + strconv.Itoa(len(photos)) + " photos found.\nPlease note that not all photos will be imported for some of them are photos fetched from a feed and are not stored locally (my guess). \nNow writing them into Day One...")
	time.Sleep(3 * time.Second)

	readyFilePath()
	writeIntoDayOne()
	writeOldPhotos()

	fmt.Println("\n\nNow say goodbye Momento.")
}

func convertTime(timeStamp time.Time) time.Time {
	return time.Unix(timeStamp.Unix()-momentoTimeStamp+unixTimeStamp, 0).Local()
}

func grabData() {
	dataPath := filePath + "/Data/backup/Database/Momento.sqlite"
	db, err := sql.Open("sqlite3", dataPath)
	if err != nil {
		fmt.Println("Can't open Momento.sqlite")
		return
	}

	// Feeds are not fetched, except RSS feeds
	rows, err := db.Query("SELECT Z_PK, ZDATE, ZNOTES, ZTITLE FROM 'ZMOMENT' WHERE ZFEEDITEMID IS NULL")
	if err != nil {
		fmt.Println(err)
		return
	}

	for rows.Next() {
		var id int64
		var createdTime time.Time
		var note sql.NullString
		var rssTitle sql.NullString
		err = rows.Scan(&id, &createdTime, &note, &rssTitle)
		if err != nil {
			fmt.Println(err)
		}
		var realNote string
		if rssTitle.Valid {
			realNote = rssTitle.String + "\n" + note.String
		} else {
			realNote = note.String
		}
		moment := &Moment{
			ID:          id,
			CreatedTime: convertTime(createdTime),
			Note:        realNote,
		}
		moments = append(moments, moment)
	}
	rows.Close()

	rows, err = db.Query("SELECT Z_PK, ZMOMENT, ZPHOTOSASSETIDENTIFIERLOCAL, ZUNIQUEIDENTIFIER FROM 'ZMEDIA'")
	if err != nil {
		fmt.Println(err)
		return
	}
	for rows.Next() {
		var id int64
		var momentId int64
		var assetIdentifierlocal sql.NullString
		var uniqueIdentifier sql.NullString
		err = rows.Scan(&id, &momentId, &assetIdentifierlocal, &uniqueIdentifier)
		if err != nil {
			fmt.Println(err)
		}
		media := &Media{
			ID:            id,
			MomentID:      momentId,
			Identifier:    uniqueIdentifier.String,
			NewInMomento3: assetIdentifierlocal.Valid,
		}
		photos = append(photos, media)
	}
	rows.Close()

	for _, moment := range moments {
		id := moment.ID
		moment.Media = photosForMoment(photos, id)
	}

	db.Close()
}

func photosForMoment(photos []*Media, id int64) (ps []*Media) {
	for _, p := range photos {
		if p.MomentID == id && p.NewInMomento3 {
			ps = append(ps, p)
		}
	}
	return
}

func readyFilePath() {
	mediaPath := filePath + "/Attachments/"
	filepath.Walk(mediaPath, walkFunc)
}

func getPhotoPath(p *Media) (path string) {
	mediaPath := filePath + "/Attachments/"
	if p.NewInMomento3 {
		path = mediaPath + p.Identifier + mediaPostfix
		return
	}
	path = mediaPath

	return
}

func writeIntoDayOne() {
	for _, m := range moments {
		media := ""
		var c2 *exec.Cmd
		if len(m.Media) > 0 {
			media = `-p=` + getPhotoPath(m.Media[0])
			c2 = exec.Command("dayone", `-d="`+m.CreatedTime.Format("01/02/2006 03:04 PM")+`"`, media, `-j=`+journalPath, "new")
		} else {
			c2 = exec.Command("dayone", `-d="`+m.CreatedTime.Format("01/02/2006 03:04 PM")+`"`, `-j=`+journalPath, "new")
		}
		c1 := exec.Command("echo", m.Note)

		r, w := io.Pipe()
		c1.Stdout = w
		c2.Stdin = r

		var b2 bytes.Buffer
		c2.Stdout = &b2

		c1.Start()
		c2.Start()
		c1.Wait()
		w.Close()
		c2.Wait()
		io.Copy(os.Stdout, &b2)

		if len(m.Media) > 20 {
			for i := 1; i < len(m.Media); i++ {
				media = "-p=" + getPhotoPath(m.Media[i])
				cmd := exec.Command("dayone", `-d="`+m.CreatedTime.Format("01/02/2006 03:04 PM")+`"`, media, `-j=`+journalPath, "new")
				cmd.Stdout = os.Stdout
				cmd.Run()
			}
		}
	}
}

func writeOldPhotos() {
	mediaPath := filePath + "/Attachments/"
	for _, imgName := range fileList {
		path := mediaPath + imgName
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		createdTime := info.ModTime().Format("01/02/2006 03:04AM")
		media := "-p=" + path
		cmd := exec.Command("dayone", `-d="`+createdTime+`"`, `-j=`+journalPath, media, "new")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func walkFunc(path string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() {
		return nil
	}

	ok, _ := regexp.MatchString("^.*IMG_\\d*?_.*$", f.Name())
	if ok {
		fileList = append(fileList, f.Name())
	}
	return nil
}
