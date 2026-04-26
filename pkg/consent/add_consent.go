package consent

import "encoding/xml"

type SoapEnvelope struct {
	XMLName xml.Name
	Body    Body
}

type Body struct {
	XMLName xml.Name
	Service Service `xml:",any"`
}

type Service struct {
	XMLName xml.Name
	Scan    Document `xml:"consent>scans,omitempty"`
}

type Document struct {
	Name string  `xml:"fileName,omitempty"`
	Type string  `xml:"fileType,omitempty"`
	Data *string `xml:"base64,omitempty"`
}
