package entre

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func newCloseNotifyingRecorder() *closeNotifyingRecorder {
	return &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func (c *closeNotifyingRecorder) close() {
	c.closed <- true
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

type hijackableResponse struct {
	Hijacked bool
}

func newHijackableResponse() *hijackableResponse {
	return &hijackableResponse{}
}

func (h *hijackableResponse) Header() http.Header {
	return nil
}

func (h *hijackableResponse) Write(buf []byte) (int, error) {
	return 0, nil
}

func (h *hijackableResponse) WriteHeader(code int) {
}

func (h *hijackableResponse) Flush() {
}

func (h *hijackableResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.Hijacked = true
	return nil, nil, nil
}

func Test_ResponseBeforeWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	r := NewResponse(rec)

	expect(t, r.Status(), 0)
	expect(t, r.Written(), false)
}

func Test_ResponseBeforeFuncHasAccessToStatus(t *testing.T) {
	var status int

	rec := httptest.NewRecorder()
	rw := NewResponse(rec)

	rw.Before(func(r Response) {
		status = r.Status()
	})
	rw.WriteHeader(http.StatusCreated)

	expect(t, status, http.StatusCreated)
}

func Test_ResponseWritingString(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponse(rec)

	rw.Write([]byte("This is our awesome new response string"))

	expect(t, rec.Code, rw.Status())
	expect(t, rec.Body.String(), "This is our awesome new response string")
	expect(t, rw.Status(), http.StatusOK)
	expect(t, rw.BodyLength(), 39)
	expect(t, rw.Written(), true)
}

func Test_ResponseWritingStrings(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponse(rec)

	rw.Write([]byte("These are our"))
	rw.Write([]byte(" awesome new response strings"))

	expect(t, rec.Code, rw.Status())
	expect(t, rec.Body.String(), "These are our awesome new response strings")
	expect(t, rw.Status(), http.StatusOK)
	expect(t, rw.BodyLength(), 42)
}

func Test_ResponseWritingHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponse(rec)

	rw.WriteHeader(http.StatusNotFound)

	expect(t, rec.Code, rw.Status())
	expect(t, rec.Body.String(), "")
	expect(t, rw.Status(), http.StatusNotFound)
	expect(t, rw.BodyLength(), 0)
}

func Test_ResponseBefore(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponse(rec)
	result := ""

	rw.Before(func(Response) {
		result += "world"
	})
	rw.Before(func(Response) {
		result += "new"
	})

	rw.WriteHeader(http.StatusNotFound)

	expect(t, rec.Code, rw.Status())
	expect(t, rec.Body.String(), "")
	expect(t, rw.Status(), http.StatusNotFound)
	expect(t, rw.BodyLength(), 0)
	expect(t, result, "newworld")
}

func Test_ResponseHijack(t *testing.T) {
	hijackable := newHijackableResponse()
	rw := NewResponse(hijackable)
	hijacker, ok := rw.(http.Hijacker)
	expect(t, ok, true)
	_, _, err := hijacker.Hijack()
	if err != nil {
		t.Error(err)
	}
	expect(t, hijackable.Hijacked, true)
}

func Test_ResponseHijackNotOK(t *testing.T) {
	hijackable := new(http.ResponseWriter)
	rw := NewResponse(*hijackable)
	hijacker, ok := rw.(http.Hijacker)
	expect(t, ok, true)
	_, _, err := hijacker.Hijack()

	refute(t, err, nil)
}

func Test_ResponseCloseNotify(t *testing.T) {
	rec := newCloseNotifyingRecorder()
	rw := NewResponse(rec)
	closed := false
	notifier := rw.(http.CloseNotifier).CloseNotify()
	rec.close()
	select {
	case <-notifier:
		closed = true
	case <-time.After(time.Second):
	}
	expect(t, closed, true)
}

func Test_ResponseNonCloseNotify(t *testing.T) {
	rw := NewResponse(httptest.NewRecorder())
	_, ok := rw.(http.CloseNotifier)
	expect(t, ok, false)
}

func TestResponseFlusher(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponse(rec)

	_, ok := rw.(http.Flusher)
	expect(t, ok, true)
}

func Test_ResponseFlushmarksWritten(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponse(rec)

	rw.Flush()
	expect(t, rw.Status(), http.StatusOK)
	expect(t, rw.Written(), true)
}
