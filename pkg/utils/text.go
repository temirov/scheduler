package utils

// WordWrap breaks a string into lines of a specified maximum width.
// It ensures that no line exceeds the given width, breaking at word
// boundaries when possible.
func WordWrap(text string, lineWidth int) []string {
	if len(text) <= lineWidth {
		return []string{text}
	}
	
	var wrappedLines []string
	var currentLine string
	words := SplitWords(text)
	
	for _, word := range words {
		// Check if word fits on current line with a space
		if len(currentLine)+len(word)+1 <= lineWidth {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			// Word doesn't fit on current line
			if currentLine != "" {
				wrappedLines = append(wrappedLines, currentLine)
			}
			
			// Handle words longer than lineWidth
			if len(word) > lineWidth {
				remainingWord := word
				for len(remainingWord) > 0 {
					if len(remainingWord) <= lineWidth {
						currentLine = remainingWord
						remainingWord = ""
					} else {
						currentLine = remainingWord[:lineWidth]
						remainingWord = remainingWord[lineWidth:]
					}
					wrappedLines = append(wrappedLines, currentLine)
					currentLine = ""
				}
			} else {
				currentLine = word
			}
		}
	}
	
	// Add the last line if there's anything left
	if currentLine != "" {
		wrappedLines = append(wrappedLines, currentLine)
	}
	
	return wrappedLines
}

// SplitWords breaks a string into individual words, considering
// spaces, tabs, and newlines as word separators.
func SplitWords(text string) []string {
	var words []string
	var currentWord string
	
	for _, character := range text {
		if character == ' ' || character == '\t' || character == '\n' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(character)
		}
	}
	
	if currentWord != "" {
		words = append(words, currentWord)
	}
	
	return words
} 