package m3u

import (
	"testing"
)

func TestParseEntry_SingleEntry(t *testing.T) {
	inputEntry := Entry{
		Metadata: "#EXTINF:-1 tvg-id=\"ShanghaiEducationTelevisionStation.cn\" tvg-logo=\"https://www.setv.sh.cn/img/logo.d54b87dd.png\" group-title=\"Culture;Education\",Shanghai Education Television Station",
	}

	_ = ParseEntry([]Entry{inputEntry})
}
