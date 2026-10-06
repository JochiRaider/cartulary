package objectstore

import "io"

// s3UploadReader admits a bounded stream and exposes no seeker/length hint.
// A terminal error prevents single-PUT publication or multipart completion.
type s3UploadReader struct {
	reader    io.Reader
	remaining int64
	exact     bool
	terminal  error
}

func (r *s3UploadReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.terminal != nil {
		return 0, r.terminal
	}
	if r.remaining == 0 {
		var extra [1]byte
		n, err := r.reader.Read(extra[:])
		if n != 0 {
			err = invalidRequest(OperationPutObject, "upload size exceeds its admitted bound")
		}
		r.terminal = err
		return 0, err
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.reader.Read(p)
	r.remaining -= int64(n)
	if err == io.EOF && r.exact && r.remaining != 0 {
		err = invalidRequest(OperationPutObject, "upload size differs from its declared length")
	}
	r.terminal = err
	return n, err
}
