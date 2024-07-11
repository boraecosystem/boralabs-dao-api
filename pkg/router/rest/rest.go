package rest

import (
	"boralabs/pkg/datastore/mongodb"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Response struct {
	*gin.Context
	BaseResponse
	Error error
}

type BaseResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	IsPaging  bool   `json:"-"`
	Paginator *mongodb.Paginator
}

type InternalServerError struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func (r *Response) setDefaults() {
	if r.Code == 0 {
		r.Code = http.StatusOK
	}

	if r.Message == "" {
		r.Message = http.StatusText(r.Code)
	}

	if r.BaseResponse.Code != 0 {
		r.Code = r.BaseResponse.Code
	}
}

func (r *Response) Json() {
	r.setDefaults()

	// error logging
	if r.Error != nil {
		log.Printf("Json Error :: %v :: %v\n", r.Error, r.Context.Request)
	}

	body := gin.H{
		"code":    r.Code,
		"message": r.Message,
		"data":    r.BaseResponse.Data,
	}
	if r.BaseResponse.IsPaging == true {
		body["paginator"] = r.BaseResponse.Paginator
	}

	r.JSON(r.Code, body)
	return
}

func (r *Response) JsonError(err error) {
	r.Error = err
	if r.Code == 0 {
		r.Code = 500
	}

	if r.Message == "" {
		r.Message = http.StatusText(r.Code)
	}
	r.Json()
}
