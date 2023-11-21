package xhttp

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
	ErrQueryParamNotFound = errors.New("query param not found")
	ErrPathParamNotFound  = errors.New("path param not found")
)

func queryParamStr(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func QueryParamStr(r *http.Request, key string) (string, error) {
	return queryParamStr(r, key), nil
}

func QueryParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	strKey := queryParamStr(r, key)
	if strKey == "" {
		return uuid.Nil, nil
	}

	return uuid.Parse(strKey)
}

func QueryParamInt(r *http.Request, key string) (int, error) {
	strKey := queryParamStr(r, key)
	if strKey == "" {
		return 0, nil
	}

	return strconv.Atoi(strKey)
}

func QueryParamBool(r *http.Request, key string) (*bool, error) {
	strKey := queryParamStr(r, key)
	if strKey == "" {
		return nil, nil
	}

	b, err := strconv.ParseBool(strKey)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func PathParamStr(r *http.Request, key string) (string, error) {
	value := mux.Vars(r)[key]
	if value == "" {
		return "", fmt.Errorf("%w: %s", ErrPathParamNotFound, key)
	}

	return value, nil
}

func PathParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	strKey, err := PathParamStr(r, key)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(strKey)
}
