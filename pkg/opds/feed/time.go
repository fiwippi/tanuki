package feed

import (
	"encoding/xml"
	"time"
)

// Time marshals time into RFC3339 in the XML
type Time struct {
	time.Time
}

func (t *Time) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(t.Format(time.RFC3339), start)
}
