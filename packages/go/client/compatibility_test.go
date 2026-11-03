package client

import "testing"

func TestExistingEnumNames(t *testing.T) {
	statuses := []QuoteBatchItemResultStatus{NOTQUOTED, PENDING, QUOTED, UNDERREVIEW}
	want := []QuoteBatchItemResultStatus{"NOT_QUOTED", "PENDING", "QUOTED", "UNDER_REVIEW"}
	for i, status := range statuses {
		if status != want[i] {
			t.Fatalf("status %d = %q", i, status)
		}
	}
	contentTypes := []UploadUrlRequestContentType{Applicationpdf, Imagejpeg, Imagepng}
	if contentTypes[0] != "application/pdf" || contentTypes[1] != "image/jpeg" || contentTypes[2] != "image/png" {
		t.Fatalf("content types changed: %v", contentTypes)
	}
	documentTypes := []UploadUrlRequestDocumentType{PROOFOFDELIVERY, SIGNATURE}
	if documentTypes[0] != "PROOF_OF_DELIVERY" || documentTypes[1] != "SIGNATURE" {
		t.Fatalf("document types changed: %v", documentTypes)
	}
}
