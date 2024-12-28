package mongodb

import (
	"fmt"
	"testing"
	"tv-server/utils/core"
)

func TestSave(t *testing.T) {
	ms := &MediaStream{
		StreamName:  "[V]音乐HD",
		ChannelName: "音乐",
		StreamUrl:   []string{"http://saka36.fansmestar.com/ch084/playlist.m3u8"},
	}
	if err := ms.Save(nil); err != nil {
		fmt.Println(err)
	}
}

func TestGetList(t *testing.T) {
	filter := QueryFilter{
		ChannelNameList: []Name{"音乐"},
	}

	c := core.NewContext()
	msList, err := filter.GetList(c)
	if err != nil {
		t.Fatal(err)
	}

	for _, ms := range msList {
		fmt.Println(ms.ID)
		fmt.Println(ms.StreamName, ms.ChannelName, ms.StreamUrl)
	}
}
