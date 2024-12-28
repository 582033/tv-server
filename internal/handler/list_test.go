package handler

import (
	"testing"
	"tv-server/utils/core"
)

func TestList(t *testing.T) {
	List(core.NewContext())
}
