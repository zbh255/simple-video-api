package main

import (
	"os"
	"testing"
)

func TestTimeFunction(t *testing.T) {
	t1, err := FFmpegTimeParse("300:59:10")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(t1)
	t.Log(FFmpegTimeFormat(t1))
}

func TestFfmpegCall(t *testing.T) {
	err := CallFfmpeg(os.Stdout,"/tmp/1f03beee-efc0-11ec-b8b8-54b203942810.mp4",
		"./testdata/video/1f03beee-efc0-11ec-b8b8-54b203942810.mp4",
		"00:00:10", "00:00:30")
	if err != nil {
		t.Fatal(err)
	}
}
