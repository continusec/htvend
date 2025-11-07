package s3store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/continusec/htvend/internal/blobstore"
)

var _ blobstore.Store = &S3Store{}

type S3StoreConfig struct {
	Bucket string
	Prefix string
}

type S3Store struct {
	config S3StoreConfig
	client *s3.Client
}

func NewS3Store(s3cfg S3StoreConfig) (*S3Store, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("err creating s3 context: %w", err)
	}
	return &S3Store{
		config: s3cfg,
		client: s3.NewFromConfig(cfg),
	}, nil
}

func (s *S3Store) keyToName(k []byte) string {
	return s.config.Prefix + hex.EncodeToString(k)
}

// Get thing with this hash
func (s *S3Store) Get(k []byte) (io.ReadCloser, error) {
	rv, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.keyToName(k)),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting object from S3: %w", err)
	}
	return rv.Body, nil
}

// Does this exist?
func (s *S3Store) Exists(k []byte) (bool, error) {
	if _, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.keyToName(k)),
	}); err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("error checking for object existence in s3: %w", err)
	}
	return true, nil
}

// Put a thing
func (s *S3Store) Put() (blobstore.ContentAddressableBlob, error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %w", err)
	}
	h := sha256.New()
	return &tbUploaded{
		s: s,
		f: f,
		h: h,
		w: io.MultiWriter(f, h),
	}, nil
}

type tbUploaded struct {
	s *S3Store
	f *os.File
	w io.Writer
	h hash.Hash
}

func (t *tbUploaded) Write(b []byte) (int, error) {
	return t.w.Write(b)
}

// Called when complete successfully. Returns hash and nil if successful.
func (t *tbUploaded) Commit() ([]byte, error) {
	if _, err := t.f.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("error seeking file start: %w", err)
	}

	k := t.h.Sum(nil)
	if _, err := t.s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(t.s.config.Bucket),
		Key:    aws.String(t.s.keyToName(k)),
		Body:   t.f,
	}); err != nil {
		return nil, fmt.Errorf("error uploading blob to s3: %w", err)
	}

	if err := t.Cleanup(); err != nil {
		return nil, fmt.Errorf("error cleaning up after upload: %w", err)
	}

	return k, nil
}

// Call if failed and should cleanup after ourselves. No-op if called after successful Commit()
func (t *tbUploaded) Cleanup() error {
	if t.f != nil {
		_ = t.f.Close() // ignore any error closing it
		if err := os.Remove(t.f.Name()); err != nil {
			return fmt.Errorf("error removing temp file for s3 upload: %w", err)
		}
		t.f = nil
	}
	return nil
}

// clean up everything - delete it all
func (s *S3Store) Destroy() error {
	return fmt.Errorf("destroy except currently not implemented for S3 blobstore implementation yet")
}

// delete everything except these (by string?)
func (s *S3Store) RemoveExcept(keep map[string]bool) error {
	return fmt.Errorf("remove except currently not implemented for S3 blobstore implementation yet")
}
