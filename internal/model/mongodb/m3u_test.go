package mongodb

import (
	"fmt"
	"testing"
)

func TestSave(t *testing.T) {
	ms := &MediaStream{
		StreamName:  "[V]音乐HD",
		ChannelName: "音乐",
		StreamUrl:   []string{"http://saka36.fansmestar.com/ch084/playlist.m3u8"},
	}
	if err := ms.Save(); err != nil {
		fmt.Println(err)
	}
}

func TestGetList(t *testing.T) {
	ms := &MediaStream{}

	filter := QueryFilter{
		ChannelNameList: []Name{"音乐"},
	}

	msList, err := ms.GetList(filter)
	if err != nil {
		t.Fatal(err)
	}

	for _, ms := range msList {
		fmt.Println(ms.ID)
		fmt.Println(ms.StreamName, ms.ChannelName, ms.StreamUrl)
	}
}
