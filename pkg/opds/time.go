package opds

import (
	"encoding/xml"
	"time"
)

type Time struct {
	time.Time
}

func (t *Time) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(t.Format(time.RFC3339), start)
}