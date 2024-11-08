package xmlnode

import (
	"encoding/xml"
	"strings"
)

// Node describes the most generic possible XML structure, for parsing
// arbitrary data and traversing it as needed, and still being able to
// re-marshal it *mostly* how it came in.
//
// Note that Content will contain all kinds of garbage since it tries to
// preserve every detail of the non-node text, such as spaces, newlines, etc.
// This is fixed upon marshaling, but if data needs to be used for anything
// beyond marshaling, users need to deal with the extraneous spacing manually.
// We could trim spaces on read instead of write, but that risks actual data
// loss in rare cases.
type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Nodes   []Node     `xml:",any"`
}

// MarshalXML implements [xml.Marshaler], cleaning preceding and trailing
// whitespace off the content to make marshaling more correct
func (n Node) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// This type alias lets us "re-marshal" our altered node, but without
	// infinite recursion that would occur if we tried to use the base type
	type node Node

	var n2 node = (node)(n)
	n2.Content = strings.TrimSpace(n2.Content)

	return e.Encode(n2)
}
