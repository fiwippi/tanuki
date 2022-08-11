package feed

import (
	"encoding/xml"
	"time"
)

// opdsTime marshals time into RFC3339 in the XML
type opdsTime struct {
	time.Time
}

func (t *opdsTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(t.Format(time.RFC3339), start)
}
