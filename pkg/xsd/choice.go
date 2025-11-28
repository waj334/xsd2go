package xsd

import (
	"encoding/xml"
	"slices"
	"strings"
)

type Choice struct {
	XMLName     xml.Name   `xml:"http://www.w3.org/2001/XMLSchema choice"`
	MinOccurs   string     `xml:"minOccurs,attr"`
	MaxOccurs   string     `xml:"maxOccurs,attr"`
	ElementList []Element  `xml:"element"`
	Sequences   []Sequence `xml:"sequence"`
	schema      *Schema
	allElements []Element
}

func (c *Choice) compile(sch *Schema, parentElement *Element) {
	c.schema = sch
	c.allElements = []Element{}

	for idx := range c.ElementList {
		el := &c.ElementList[idx]

		el.compile(sch, parentElement)
		// Propagate array cardinality downwards
		if c.MaxOccurs == "unbounded" {
			el.MaxOccurs = "unbounded"
		}
		if el.MinOccurs == "" {
			el.MinOccurs = "0"
		}
	}

	c.allElements = append(c.allElements, c.ElementList...)

	for i := 0; i < len(c.allElements); i++ {
		el := c.allElements[i]
		elements, ok := sch.substitutionGroup[el.Ref]
		if ok {
			for _, subst := range elements {
				if _, imported := subst.schema.importedModules[el.Ref.NsPrefix()]; imported {
					// Prevent creating a circular dependency.
					continue
				}

				prefix := subst.schema.Xmlns.PrefixByUri(subst.schema.TargetNamespace)
				ref := reference(prefix + ":" + subst.Name)
				substEl := Element{
					Ref:       ref,
					MinOccurs: el.MinOccurs,
					MaxOccurs: el.MaxOccurs,
				}
				substEl.compile(sch, parentElement)
				c.allElements = append(c.allElements, substEl)
			}
		}
	}

	inheritedElements := []Element{}
	for idx := range c.Sequences {
		el := &c.Sequences[idx]
		el.compile(sch, parentElement)
		for _, el2 := range el.Elements() {
			if c.MaxOccurs == "unbounded" {
				el2.MaxOccurs = "unbounded"
			}
			if el2.MinOccurs == "" {
				el2.MinOccurs = "0"
			}
			inheritedElements = append(inheritedElements, el2)
		}
	}
	// deduplicate elements that represent duplicate within xsd:choice/xsd:sequence structure
	c.allElements = append(c.allElements, deduplicateElements(inheritedElements)...)
	c.allElements = deduplicateElements(c.allElements)
	slices.SortFunc(c.allElements, func(a, b Element) int {
		return strings.Compare(a.GoName(), b.GoName())
	})
}

func (c *Choice) Elements() []Element {
	return c.allElements
}
