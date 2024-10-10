package notifihttp_test

// import (
// 	"bytes"
// 	"context"
// 	"net/http"
// 	"testing"

// 	"github.com/stevenferrer/notifi/notifihttp"
// 	"github.com/stretchr/testify/assert"
// )

// func TestRequestSender(t *testing.T) {
// 	rs := notifihttp.NewDefaultRequestSender().
// 		WithHTTPClient(http.DefaultClient)

// 	buf := &bytes.Buffer{}
// 	buf.Write([]byte("hello"))
// 	urlStr := "https://webhook.site/3d597059-4092-4cc3-8c00-b988a6e7bf83"
// 	err := rs.SendRequest(context.Background(), urlStr, buf)
// 	assert.NoError(t, err)
// }
