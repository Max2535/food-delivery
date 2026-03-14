package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

func main() {}

// HandlerRegisterer is the symbol KrakenD looks for when loading the plugin
var HandlerRegisterer = registerer("correlation-id-injector")

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.handler)
}

func (r registerer) handler(_ context.Context, _ map[string]interface{}, next http.Handler) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// สร้าง Correlation ID ถ้า Client ไม่ส่งมา
		if req.Header.Get("X-Correlation-ID") == "" {
			req.Header.Set("X-Correlation-ID", newUUID())
		}
		next.ServeHTTP(w, req)
	}), nil
}

// newUUID สร้าง UUID v4 ด้วย crypto/rand (cryptographically secure)
func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
