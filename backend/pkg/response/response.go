package response

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Ok(w http.ResponseWriter) {
	httpx.OkJson(w, Body{Code: 0, Msg: "success"})
}

func OkWithData(w http.ResponseWriter, data interface{}) {
	httpx.OkJson(w, Body{Code: 0, Msg: "success", Data: data})
}

func Error(w http.ResponseWriter, code int, msg string) {
	httpx.OkJson(w, Body{Code: code, Msg: msg})
}

func Unauthorized(w http.ResponseWriter, msg string) {
	httpx.WriteJson(w, http.StatusUnauthorized, Body{Code: 401, Msg: msg})
}
