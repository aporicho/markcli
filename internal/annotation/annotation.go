package annotation

type Annotation struct {
	ID           string `json:"id"`
	StartLine    int    `json:"startLine"`
	EndLine      int    `json:"endLine"`
	StartCol     *int   `json:"startCol,omitempty"`
	EndCol       *int   `json:"endCol,omitempty"`
	SelectedText string `json:"selectedText"`
	Comment      string `json:"comment"`
	CreatedAt    string `json:"createdAt"`
	Resolved     *bool  `json:"resolved,omitempty"`
	Quote        string `json:"quote,omitempty"`
	Prefix       string `json:"prefix,omitempty"`
	Suffix       string `json:"suffix,omitempty"`
}

type AnnotationFile struct {
	File        string       `json:"file"`
	Annotations []Annotation `json:"annotations"`
}

type SelectionPos struct {
	Line int // 1-based
	Col  int // 0-based
}

func IsResolved(a Annotation) bool {
	return a.Resolved != nil && *a.Resolved
}
