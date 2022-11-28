package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test404BasicHandler(t *testing.T) {
	tServer := httptest.NewServer(getHandler())
	defer tServer.Close()

	res, err := tServer.Client().Get(tServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestPostHash400hWithoutHeaderPassHandler(t *testing.T) {
	tServer := httptest.NewServer(getHandler())
	defer tServer.Close()

	url := fmt.Sprintf("%s/hash", tServer.URL)

	res, err := tServer.Client().Post(url, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestPostHash400hWithWrongPathHandler(t *testing.T) {
	tServer := httptest.NewServer(getHandler())
	defer tServer.Close()

	url := fmt.Sprintf("%s/hash/id/pass", tServer.URL)

	res, err := tServer.Client().Post(url, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestPostHash400hWithEmptyHeaderPassHandler(t *testing.T) {
	tServer := httptest.NewServer(getHandler())
	defer tServer.Close()

	url := fmt.Sprintf("%s/hash", tServer.URL)

	httptest.NewRecorder().Header().Add("password", "")
	res, err := tServer.Client().Post(url, "", nil)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestPostHash200hWithEmptyHeaderPassHandler(t *testing.T) {
	tServer := httptest.NewServer(getHandler())
	defer tServer.Close()

	req := httptest.NewRequest(http.MethodPost, "/hash", nil)
	req.Header.Set("password", "qwerty")
	rest := httptest.NewRecorder()

	postHash(rest, req)

	res := rest.Result()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "1", string(body))
}

func Test200GetHashHandler(t *testing.T) {
	go startEncodePass(passMapChan)

	tServer := httptest.NewServer(getHandler())
	defer tServer.Close()

	req := httptest.NewRequest(http.MethodPost, "/hash", nil)
	req.Header.Set("password", "qwerty")
	rest := httptest.NewRecorder()

	postHash(rest, req)
	res := rest.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	res.Body.Close()

	_, err = strconv.Atoi(string(body))
	if err != nil {
		t.Error(err)
	}

	url := fmt.Sprintf("%s/hash/%s", tServer.URL, body)

	resGet, err := tServer.Client().Get(url)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusAccepted, resGet.StatusCode)

	time.Sleep(5 * time.Second)

	resGet, err = tServer.Client().Get(url)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resGet.StatusCode)
}
