package handler

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func Success(ctx *fasthttp.RequestCtx, data any) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	if body, err := json.Marshal(data); err == nil {
		ctx.SetBody(body)
	} else {
		ctx.Error("Failed to encode response", fasthttp.StatusInternalServerError)
	}
}

func ErrorWithMessage(ctx *fasthttp.RequestCtx, status int, message string) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	response := map[string]any{"error": message}
	if body, err := json.Marshal(response); err == nil {
		ctx.SetBody(body)
	} else {
		ctx.Error("Failed to encode response", fasthttp.StatusInternalServerError)
	}
}
