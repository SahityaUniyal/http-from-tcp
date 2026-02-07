package request

import (
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}

	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))

	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	return n, nil
}

func TestRequestFromReader(t *testing.T) {
	tests := []struct {
		name    string
		reader  io.Reader
		want    *Request
		wantErr bool
	}{
		{
			name: "Good GET Request Line",
			reader: &chunkReader{
				data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
				numBytesPerRead: 3,
			},
			want: &Request{
				RequestLine: RequestLine{
					Method:        "GET",
					RequestTarget: "/",
					HttpVersion:   "1.1",
				},
				State: 2,
			},
			wantErr: false,
		},
		{
			name: "Good GET Request Line With Path",
			reader: &chunkReader{
				data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
				numBytesPerRead: 1,
			},
			want: &Request{
				RequestLine: RequestLine{
					Method:        "GET",
					RequestTarget: "/coffee",
					HttpVersion:   "1.1",
				},
				State: 2,
			},
			wantErr: false,
		},
		{
			name: "Invalid number of parts in request line",
			reader: &chunkReader{
				data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
				numBytesPerRead: 3,
			},
			wantErr: true,
		},
		{
			name: "Invalid method",
			reader: &chunkReader{
				data:            "WRONG /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
				numBytesPerRead: 3,
			},
			wantErr: true,
		},
		{
			name: "Invalid http version",
			reader: &chunkReader{
				data:            "WRONG /coffee HTTP/2.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
				numBytesPerRead: 3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := RequestFromReader(tt.reader)

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			if tt.want != nil && !reflect.DeepEqual(got.RequestLine, tt.want.RequestLine) {
				t.Errorf("RequestFromReader() response not equal got:%v \n want: %v", got, tt.want)

			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}
