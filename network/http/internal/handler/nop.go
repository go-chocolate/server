package handler

import "net/http"

var NopHandler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {

}
