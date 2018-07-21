package tcpchecker

import (
	"log"
	"testing"
	"time"
)

func TestDown(t *testing.T) {
	checker, err := New(Option{
		DefaultDown: false,
	})
	if err != nil {
		log.Fatal("init failed", err)
	}
	if res := checker.Down("hui.lu:81"); res {
		t.Error("it should be false", res)
	}
	if res := checker.Down("hui.lu:80"); res {
		t.Error("it should be false", res)
	}
	checker.AddRef("hui.lu:443")
	time.Sleep((defaultCheckInterval + defaultcheckTimeout) * defaultFail)

	if res := checker.Down("hui.lu:81"); !res {
		t.Error("it should bt true", res)
	}
	if res := checker.Down("hui.lu:80"); res {
		t.Error("it should bt false", res)
	}
	if res := checker.Down("hui.lu:443"); res {
		t.Error("it should bt false", res)
	}
}
