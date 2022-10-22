package reqlogs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/loebfly/ezgin/internal/engine/middleware/trace"
	"github.com/loebfly/ezgin/internal/logs"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func (receiver enter) Middleware(c *gin.Context) {
	// 不记录静态文件和根目录请求
	if strings.Contains(c.Request.RequestURI, "/docs") || c.Request.RequestURI == "/" {
		return
	}

	rWriter := &respWriter{
		body:           bytes.NewBufferString(""),
		ResponseWriter: c.Writer,
	}
	c.Writer = rWriter

	// 开始时间
	startTime := time.Now()
	reqTime := startTime.Format("2006-01-02 15:04:05.012")

	rawData, err := c.GetRawData()
	if err != nil {
		logs.Enter.CError("GIN", "GetRawData error:{}", err.Error())
	}
	var reqHeaders = c.Request.Header.Clone()
	// 关键点 重置请求体
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(rawData))

	// 处理请求
	c.Next()

	var reqParams interface{}
	if strings.Contains(c.ContentType(), gin.MIMEJSON) {
		var params = make(map[string]interface{})
		err = json.Unmarshal(rawData, &reqParams)
		if err != nil {
			logs.Enter.CError("GIN", "reqParams json.unmarshal error:{}", err.Error())
		}
		reqParams = params
	} else if strings.Contains(c.ContentType(), gin.MIMEPOSTForm) ||
		strings.Contains(c.ContentType(), gin.MIMEMultipartPOSTForm) {
		reqParams = receiver.GetFormParams(c)
	} else {
		reqParams = string(rawData)
	}

	endTime := time.Now()
	respTime := endTime.Format("2006-01-02 15:04:05.012")

	var respParams = make(map[string]interface{})
	respStr := rWriter.body.String()
	if respStr != "" && respStr[0:1] == "{" {
		err = json.Unmarshal(rWriter.body.Bytes(), &respParams)
		if err != nil {
			logs.Enter.CError("GIN", "respParams json.Unmarshal error:{}", err.Error())
		}
	}

	ttl := int(endTime.UnixNano()/1e6 - startTime.UnixNano()/1e6)

	method := c.Request.Method
	contentType := c.ContentType()
	uri := c.Request.RequestURI

	logs.Enter.CDebug("GIN", "|{}|{}|{}|{}|{}ms", method, uri, c.ClientIP(), respTime, ttl)
	if reqHeaders != nil {
		logs.Enter.CDebug("GIN", "请求头:{}", reqHeaders)
	}
	if reqParams != nil {
		if receiver.argToString(reqParams) != "" {
			logs.Enter.CDebug("GIN", "请求参数:"+receiver.argToString(reqParams))
		}
	}

	logs.Enter.CDebug("GIN", "响应结果:{}", respParams)

	ctx := ReqCtx{
		RequestId:   trace.Enter.GetCurReqId(),
		ReqTime:     reqTime,
		ReqHeaders:  reqHeaders,
		ReqParams:   reqParams,
		RespTime:    respTime,
		RespParams:  respParams,
		TTL:         ttl,
		Method:      method,
		ContentType: contentType,
		URI:         uri,
	}
	logChan <- ctx
}

func (receiver enter) GetFormParams(ctx *gin.Context) map[string]string {
	params := make(map[string]string)
	cType := ctx.ContentType()
	if !strings.Contains(cType, gin.MIMEPOSTForm) &&
		!strings.Contains(cType, gin.MIMEMultipartPOSTForm) {
		return params
	}
	if ctx.Request == nil {
		return params
	}
	if ctx.Request.Method == "GET" {
		for k, v := range ctx.Request.URL.Query() {
			params[k] = v[0]
		}
		return params
	} else {
		err := ctx.Request.ParseForm()
		if err != nil {
			return params
		}
		for k, v := range ctx.Request.PostForm {
			params[k] = v[0]
		}
		for k, v := range ctx.Request.URL.Query() {
			params[k] = v[0]
		}
		return params
	}
}

// ConvToString 任意类型转换为字符串
func (receiver enter) argToString(iFace interface{}) string {
	switch val := iFace.(type) {
	case []byte:
		return string(val)
	case string:
		return val
	}
	v := reflect.ValueOf(iFace)
	switch v.Kind() {
	case reflect.Invalid:
		return ""
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Ptr, reflect.Struct, reflect.Map, reflect.Array, reflect.Slice:
		b, err := json.Marshal(v.Interface())
		if err != nil {
			return ""
		}
		str := string(b)
		if v.Kind() == reflect.Map && str == "{}" {
			return "{ }"
		}
		return str
	}
	return fmt.Sprintf("%v", iFace)
}
