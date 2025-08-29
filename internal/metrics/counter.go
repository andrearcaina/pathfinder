package metrics

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

func CountFile(path string, bufferSize int, commentType CommentType) (
	code int,
	comments int,
	blanks int,
	ann AnnotationMetrics,
) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	return scanLines(f, bufferSize, commentType)
}

func scanLines(r io.Reader, bufferSize int, commentType CommentType) (
	code int,
	comments int,
	blanks int,
	ann AnnotationMetrics,
) {
	br := bufio.NewReaderSize(r, bufferSize)
	inBlockComment := false

	for {
		line, err := br.ReadBytes('\n')
		if len(line) == 0 && err != nil {
			break
		}

		line = bytes.TrimSpace(line)

		if len(line) == 0 {
			blanks++
		} else if len(commentType.SingleLine) > 0 && bytes.HasPrefix(line, []byte(commentType.SingleLine)) {
			comments++
			checkAnnotationsBytes(line, &ann)
		} else if len(commentType.BlockStart) > 0 && bytes.Contains(line, []byte(commentType.BlockStart)) {
			i := bytes.Index(line, []byte(commentType.BlockStart))
			before := bytes.TrimSpace(line[:i])
			after := line[i+len(commentType.BlockStart):]

			if len(commentType.BlockEnd) > 0 && bytes.Contains(after, []byte(commentType.BlockEnd)) {
				j := bytes.Index(after, []byte(commentType.BlockEnd))
				left := before
				right := bytes.TrimSpace(after[j+len(commentType.BlockEnd):])

				if len(left) > 0 || len(right) > 0 {
					code++
				} else {
					comments++
				}
			} else {
				if len(before) > 0 {
					code++
				} else {
					comments++
				}
				inBlockComment = true
			}

			checkAnnotationsBytes(line, &ann)
		} else if inBlockComment {
			comments++
			if len(commentType.BlockEnd) > 0 && bytes.Contains(line, []byte(commentType.BlockEnd)) {
				inBlockComment = false
			}
		} else {
			code++
		}

		if err == io.EOF {
			break
		}
	}

	return
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
