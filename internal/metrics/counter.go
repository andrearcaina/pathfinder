package metrics

import (
	"bufio"
	"io"
	"os"
	"strings"
)

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB = 1 << (10 * iota)

	maxBufferSize = 1 << 16 // 2^16 bytes (64 KB)
)

func CountFile(path string, commentType CommentType) (code int, comments int, blanks int, ann AnnotationMetrics) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	return scanLines(f, commentType)
}

func scanLines(r io.Reader, commentType CommentType) (code int, comments int, blanks int, ann AnnotationMetrics) {
	sc := bufio.NewScanner(r)

	buf := make([]byte, 0, maxBufferSize)
	sc.Buffer(buf, MB)

	inBlockComment := false

	for sc.Scan() {
		line := sc.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			blanks++
			continue
		}

		if commentType.SingleLine != "" && strings.HasPrefix(line, commentType.SingleLine) {
			comments++
			checkAnnotations(line, &ann)
			continue
		}

		if commentType.BLockStart != "" {
			if i := strings.Index(line, commentType.BLockStart); i != -1 {
				beforeIndex := strings.TrimSpace(line[:i])
				afterIndex := line[i+len(commentType.BLockStart):]

				if commentType.BlockEnd != "" && strings.Contains(afterIndex, commentType.BlockEnd) {
					j := strings.Index(afterIndex, commentType.BlockEnd)
					left := beforeIndex
					right := strings.TrimSpace(afterIndex[j+len(commentType.BlockEnd):])
					if left != "" || right != "" {
						code++
					} else {
						comments++
					}
				} else {
					if beforeIndex != "" {
						code++
					} else {
						comments++
					}
					inBlockComment = true
				}

				checkAnnotations(line, &ann)
				continue
			}
		}

		if inBlockComment {
			comments++
			if commentType.BlockEnd != "" && strings.Contains(line, commentType.BlockEnd) {
				inBlockComment = false
			}
			continue
		}

		code++
	}
	return
}

func checkAnnotations(line string, ann *AnnotationMetrics) {
	switch {
	case strings.Contains(line, "TODO"):
		ann.TotalTODO++
	case strings.Contains(line, "FIXME"):
		ann.TotalFIXME++
	case strings.Contains(line, "HACK"):
		ann.TotalHACK++
	}
	ann.TotalAnnotations = ann.TotalTODO + ann.TotalFIXME + ann.TotalHACK
}
