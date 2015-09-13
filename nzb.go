package nzb

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// Sort NzbFile by part number.
type NzbByPart []*NzbFile

func (s NzbByPart) Len() int           { return len(s) }
func (s NzbByPart) Less(i, j int) bool { return s[i].Part < s[j].Part }
func (s NzbByPart) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type Nzb struct {
	Meta  map[string]string
	Files []*NzbFile
}

func NewString(data string) (*Nzb, error) {
	return New(bytes.NewBufferString(data))
}

func New(buf io.Reader) (*Nzb, error) {
	xnzb := new(xNzb)
	dec := xml.NewDecoder(buf)
	// TODO(wathiede): some NZB's are iso-8859-1, if we need to support more,
	// handled here.
	dec.CharsetReader = func(charset string, r io.Reader) (io.Reader, error) {
		switch charset {
		case "iso-8859-1", "utf-8":
			return r, nil
		}
		return nil, fmt.Errorf("No encoding translator for %q", charset)
	}
	if err := dec.Decode(xnzb); err != nil {
		return nil, err
	}
	// convert to nicer format
	nzb := new(Nzb)
	// convert metadata
	nzb.Meta = make(map[string]string)
	for _, md := range xnzb.Metadata {
		nzb.Meta[md.Type] = md.Value
	}
	for i, _ := range xnzb.File {
		nzb.Files = append(nzb.Files, &xnzb.File[i])
	}
	return nzb, nil
}

// used only for unmarshalling xml
type xNzb struct {
	XMLName  xml.Name   `xml:"nzb"`
	Metadata []xNzbMeta `xml:"head>meta"`
	File     []NzbFile  `xml:"file"` // xml:tag name doesn't work?
}

// used only in unmarshalling xml
type xNzbMeta struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",innerxml"`
}

type NzbFile struct {
	Groups   []string     `xml:"groups>group"`
	Segments []NzbSegment `xml:"segments>segment"`
	Poster   string       `xml:"poster,attr"`
	Date     int          `xml:"date,attr"`
	Subject  string       `xml:"subject,attr"`
	Part     int
}

type NzbSegment struct {
	XMLName xml.Name `xml:"segment"`
	Bytes   int      `xml:"bytes,attr"`
	Number  int      `xml:"number,attr"`
	Id      string   `xml:",innerxml"`
}
