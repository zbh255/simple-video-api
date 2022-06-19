package main

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func CallFfmpeg(out io.Writer,videoPath string,videoOutName string,start,end string) error {
	cmd := exec.Command("ffmpeg",
		"-i",videoPath, "-ss",start, "-t",end,
		"-codec","copy",videoOutName)
	cmd.Stdout = out
	return cmd.Run()
}

func FFmpegTimeParse(str string) (time.Time,error) {
	timeSplit := strings.SplitN(str,":",3)
	if len(timeSplit) != 3 {
		return time.Time{},errors.New("time format error")
	}
	t := time.Time{}
	hourN,err := strconv.ParseInt(timeSplit[0],10,64)
	if err != nil {
		return t,err
	}
	t = t.Add(time.Hour * time.Duration(hourN))
	minuteN,err := strconv.ParseInt(timeSplit[1],10,64)
	if err != nil {
		return t,err
	}
	t = t.Add(time.Minute * time.Duration(minuteN))
	secondN,err := strconv.ParseInt(timeSplit[2],10,64)
	if err != nil {
		return t,err
	}
	t = t.Add(time.Second * time.Duration(secondN))
	return t,nil
}

func FFmpegTimeFormat(t time.Time) string {
	var sb strings.Builder
	var nHour int
	nHour += (t.Day() - 1) * 24
	nHour += t.Hour()
	sb.WriteString(strconv.Itoa(nHour))
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(t.Minute()))
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(t.Second()))
	return sb.String()
}

func GetVideoFromUrl(url string,videoName string) error {
	file, err := os.OpenFile("/tmp/"+videoName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	rep,err := http.Get(url)
	if err != nil {
		return err
	}
	defer rep.Body.Close()
	// data type is gzip
	if rep.Header.Get("Content-Type") == "encoding/gzip" {
		gReader, err := gzip.NewReader(rep.Body)
		if err != nil {
			return err
		}
		defer gReader.Close()
		_,err = io.Copy(file,gReader)
		return err
	}
	_, err = io.Copy(file,rep.Body)
	if err != nil {
		return err
	}
	return nil
}

func GetVideoFile(videoName string) (*os.File,error) {
	return os.Open("/tmp/" + videoName)
}