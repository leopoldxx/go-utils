package httputils_test

import (
	"net/url"
	"testing"

	. "github.com/leopoldxx/go-utils/httputils"
)

func TestPackURL(t *testing.T) {
	addr := "http://www.baidu.com?test=true"
	query := url.Values{
		"app_id": []string{"1234567"},
		"jumpto": []string{"http://www.google.com"},
	}
	u, err := PackURL(addr, query)
	if err != nil {
		t.Fatal("pack url failed: ", err)
	}
	t.Log(u)
}
