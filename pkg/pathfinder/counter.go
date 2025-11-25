package pathfinder

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

func fileCounter(path string, bufferSize int, langDef *LanguageDefinition) (LanguageMetrics, AnnotationMetrics, error) {
	f, err := os.Open(path)
	if err != nil {
		return LanguageMetrics{}, AnnotationMetrics{}, err
	}
	defer f.Close()

	langMetrics, annMetrics, err := countLinesInFile(f, bufferSize, langDef)
	if err != nil {
		return LanguageMetrics{}, AnnotationMetrics{}, err
	}

	return langMetrics, annMetrics, nil
}

func countLinesInFile(r io.Reader, bufferSize int, langDef *LanguageDefinition) (LanguageMetrics, AnnotationMetrics, error) {
	br := bufio.NewReaderSize(r, bufferSize)
	inBlockComment := false

	var langMetrics LanguageMetrics
	var annMetrics AnnotationMetrics

	for {
		line, err := br.ReadBytes('\n')
		if len(line) == 0 && err != nil {
			break
		}

		line = bytes.TrimSpace(line)

		if len(line) == 0 {
			langMetrics.Blanks++
		} else if len(langDef.Type.SingleLine) > 0 && bytes.HasPrefix(line, []byte(langDef.Type.SingleLine)) {
			langMetrics.Comments++
			checkAnnotationsBytes(line, &annMetrics)
		} else if len(langDef.Type.BlockStart) > 0 && bytes.Contains(line, []byte(langDef.Type.BlockStart)) {
			i := bytes.Index(line, []byte(langDef.Type.BlockStart))
			before := bytes.TrimSpace(line[:i])
			after := line[i+len(langDef.Type.BlockStart):]

			if len(langDef.Type.BlockEnd) > 0 && bytes.Contains(after, []byte(langDef.Type.BlockEnd)) {
				j := bytes.Index(after, []byte(langDef.Type.BlockEnd))
				left := before
				right := bytes.TrimSpace(after[j+len(langDef.Type.BlockEnd):])

				if len(left) > 0 || len(right) > 0 {
					langMetrics.Code++
				} else {
					langMetrics.Comments++
				}
			} else {
				if len(before) > 0 {
					langMetrics.Code++
				} else {
					langMetrics.Comments++
				}
				inBlockComment = true
			}

			checkAnnotationsBytes(line, &annMetrics)
		} else if inBlockComment {
			langMetrics.Comments++
			if len(langDef.Type.BlockEnd) > 0 && bytes.Contains(line, []byte(langDef.Type.BlockEnd)) {
				inBlockComment = false
			}
		} else {
			langMetrics.Code++
		}

		if err == io.EOF {
			break
		}
	}

	langMetrics.Language = langDef.Name
	langMetrics.Files = 1
	langMetrics.Lines = langMetrics.Code + langMetrics.Comments + langMetrics.Blanks

	/* example output
	langMetrics = LanguageMetrics{ Language: "Go", Files: 1, Code: 100, Comments: 20, Blanks: 10, Lines: 130 }
	annMetrics = AnnotationMetrics{ TotalTODO: 5, TotalFIXME: 2, TotalHACK: 1, TotalAnnotations: 8 }
	error = nil
	*/
	return langMetrics, annMetrics, nil
}

func checkAnnotationsBytes(line []byte, ann *AnnotationMetrics) {
	switch {
	case bytes.Contains(line, []byte("TODO")):
		ann.TotalTODO++
	case bytes.Contains(line, []byte("FIXME")):
		ann.TotalFIXME++
	case bytes.Contains(line, []byte("HACK")):
		ann.TotalHACK++
	}
	ann.TotalAnnotations = ann.TotalTODO + ann.TotalFIXME + ann.TotalHACK
}
