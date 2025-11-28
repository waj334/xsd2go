package xsd

import (
	"encoding/xml"
	"slices"
	"strings"
)

type Sequence struct {
	XMLName     xml.Name  `xml:"http://www.w3.org/2001/XMLSchema sequence"`
	ElementList []Element `xml:"element"`
	Choices     []Choice  `xml:"choice"`
	allElements []Element
}

func (s *Sequence) Elements() []Element {
	return s.allElements
}

func (s *Sequence) compile(sch *Schema, parentElement *Element) {
	s.allElements = []Element{}

	for idx := range s.ElementList {
		el := &s.ElementList[idx]
		el.compile(sch, parentElement)
	}
	s.allElements = s.ElementList

	for i := 0; i < len(s.allElements); i++ {
		el := s.allElements[i]
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
				s.allElements = append(s.allElements, substEl)
			}
		}
	}

	for idx := range s.Choices {
		c := &s.Choices[idx]
		c.compile(sch, parentElement)

		s.allElements = append(s.allElements, c.Elements()...)
	}

	s.allElements = deduplicateElements(s.allElements)
	slices.SortFunc(s.allElements, func(a, b Element) int {
		return strings.Compare(a.GoName(), b.GoName())
	})
}

type SequenceAll struct {
	XMLName     xml.Name  `xml:"http://www.w3.org/2001/XMLSchema all"`
	ElementList []Element `xml:"element"`
	Choices     []Choice  `xml:"choice"`
	allElements []Element
}

func (s *SequenceAll) Elements() []Element {
	return s.allElements
}

func (s *SequenceAll) compile(sch *Schema, parentElement *Element) {
	s.allElements = []Element{}

	for idx := range s.ElementList {
		el := &s.ElementList[idx]
		el.compile(sch, parentElement)

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
				s.allElements = append(s.allElements, substEl)
			}
		}
	}
	s.allElements = append(s.allElements, s.ElementList...)

	for idx := range s.Choices {
		c := &s.Choices[idx]
		c.compile(sch, parentElement)

		s.allElements = append(s.allElements, c.Elements()...)
	}

	s.allElements = deduplicateElements(s.allElements)
	slices.SortFunc(s.allElements, func(a, b Element) int {
		return strings.Compare(a.GoName(), b.GoName())
	})
}
