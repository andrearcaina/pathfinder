package metrics

var (
	languageDefinitions = []LanguageDefinition{
		{
			Name: "Go",
			Type: slashType(),
			Ext:  []string{".go"},
		},
		{
			Name: "JavaScript",
			Type: slashType(),
			Ext:  []string{".js", ".jsx", ".mjs"},
		},
		{
			Name: "TypeScript",
			Type: slashType(),
			Ext:  []string{".ts", ".tsx", ".mts"},
		},
		{
			Name: "PHP",
			Type: slashType(),
			Ext:  []string{".php"},
		},
		{
			Name: "HTML",
			Type: htmlType(),
			Ext:  []string{".html", ".htm"},
		},
		{
			Name: "CSS",
			Type: slashType(),
			Ext:  []string{".css", ".scss", ".sass"},
		},
		{
			Name: "XML",
			Type: htmlType(),
			Ext:  []string{".xml"},
		},
		{
			Name: "Python",
			Type: hashType(),
			Ext:  []string{".py", ".pyc"},
		},
		{
			Name: "Java",
			Type: slashType(),
			Ext:  []string{".java"},
		},
		{
			Name: "Kotlin",
			Type: slashType(),
			Ext:  []string{".kt", ".kts"},
		},
		{
			Name: "C",
			Type: slashType(),
			Ext:  []string{".c", ".h"},
		},
		{
			Name: "C++",
			Type: slashType(),
			Ext:  []string{".cpp", ".hpp"},
		},
		{
			Name: "C#",
			Type: slashType(),
			Ext:  []string{".cs"},
		},
		{
			Name: "Swift",
			Type: slashType(),
			Ext:  []string{".swift"},
		},
		{
			Name: "JSON",
			Type: noneType(),
			Ext:  []string{".json"},
		},
		{
			Name: "YAML",
			Type: hashType(),
			Ext:  []string{".yaml", ".yml"},
		},
		{
			Name: "Markdown",
			Type: hashType(),
			Ext:  []string{".md"},
		},
	}

	languageByExt = make(map[string]*LanguageDefinition)
)

func init() { // this function initializes the languageByExt map by iterating over the languageDefinitions slice
	for i := range languageDefinitions {
		for _, ext := range languageDefinitions[i].Ext {
			languageByExt[ext] = &languageDefinitions[i]
		}
	}
}

func DetermineLangByExt(ext string) *LanguageDefinition {
	if lang, ok := languageByExt[ext]; ok {
		return lang
	}
	return nil
}

func slashType() CommentType {
	return CommentType{
		SingleLine: "//",
		BLockStart: "/*",
		BlockEnd:   "*/",
	}
}

func htmlType() CommentType {
	return CommentType{
		BLockStart: "<!--",
		BlockEnd:   "-->",
	}
}

func hashType() CommentType {
	return CommentType{
		SingleLine: "#",
	}
}

func dashType() CommentType {
	return CommentType{
		SingleLine: "--",
	}
}

func noneType() CommentType {
	return CommentType{}
}
