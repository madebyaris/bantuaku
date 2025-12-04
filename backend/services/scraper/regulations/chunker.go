package regulations

import (
	"strings"
	"unicode"
)

// Chunker handles semantic chunking of regulation text
type Chunker struct {
	chunkSize    int // Target chunk size in characters
	overlapSize  int // Overlap size between chunks
	minChunkSize int // Minimum chunk size
}

// NewChunker creates a new chunker with default settings
func NewChunker() *Chunker {
	return &Chunker{
		chunkSize:    1000,  // ~1000 characters per chunk
		overlapSize:  200,   // 200 character overlap
		minChunkSize: 200,   // Minimum chunk size
	}
}

// Chunk splits text into semantic chunks with overlap
type Chunk struct {
	Text           string
	Index          int
	StartCharOffset int
	EndCharOffset   int
}

// ChunkText splits text into semantic chunks
func (c *Chunker) ChunkText(text string) []Chunk {
	var chunks []Chunk
	
	if len(text) <= c.chunkSize {
		// Text fits in one chunk
		return []Chunk{{
			Text:           text,
			Index:          0,
			StartCharOffset: 0,
			EndCharOffset:  len(text),
		}}
	}

	// Split by sections first (if available)
	sections := c.splitBySections(text)
	
	currentIndex := 0
	for _, section := range sections {
		if len(section) <= c.chunkSize {
			// Section fits in one chunk
			chunks = append(chunks, Chunk{
				Text:           section,
				Index:          currentIndex,
				StartCharOffset: currentIndex * c.chunkSize,
				EndCharOffset:  currentIndex*c.chunkSize + len(section),
			})
			currentIndex++
		} else {
			// Split section into smaller chunks
			sectionChunks := c.splitSection(section, currentIndex)
			chunks = append(chunks, sectionChunks...)
			currentIndex += len(sectionChunks)
		}
	}

	return chunks
}

// splitBySections splits text by section markers (Pasal, Bab, etc.)
func (c *Chunker) splitBySections(text string) []string {
	var sections []string
	var currentSection strings.Builder

	lines := strings.Split(text, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line is a section marker
		if c.isSectionMarker(line) {
			// Save current section if it has content
			if currentSection.Len() > 0 {
				sections = append(sections, currentSection.String())
				currentSection.Reset()
			}
		}

		currentSection.WriteString(line)
		currentSection.WriteString("\n")
	}

	// Add last section
	if currentSection.Len() > 0 {
		sections = append(sections, currentSection.String())
	}

	// If no sections found, return entire text as one section
	if len(sections) == 0 {
		return []string{text}
	}

	return sections
}

// isSectionMarker checks if a line is a section marker
func (c *Chunker) isSectionMarker(line string) bool {
	lineUpper := strings.ToUpper(line)
	
	// Common Indonesian regulation section markers
	markers := []string{
		"BAB", "PASAL", "AYAT", "BAGIAN", "PARAGRAF",
		"PENJELASAN", "LAMPIRAN",
	}
	
	for _, marker := range markers {
		if strings.HasPrefix(lineUpper, marker) {
			return true
		}
	}
	
	return false
}

// splitSection splits a section into chunks with overlap
func (c *Chunker) splitSection(section string, startIndex int) []Chunk {
	var chunks []Chunk
	
	// Try to split at sentence boundaries
	sentences := c.splitIntoSentences(section)
	
	var currentChunk strings.Builder
	currentLength := 0
	chunkIndex := startIndex
	startOffset := 0

	for _, sentence := range sentences {
		sentenceLen := len(sentence)
		
		if currentLength+sentenceLen > c.chunkSize && currentLength > c.minChunkSize {
			// Save current chunk
			chunkText := currentChunk.String()
			chunks = append(chunks, Chunk{
				Text:           chunkText,
				Index:          chunkIndex,
				StartCharOffset: startOffset,
				EndCharOffset:  startOffset + len(chunkText),
			})
			
			// Start new chunk with overlap
			overlapText := c.getOverlapText(chunkText, c.overlapSize)
			currentChunk.Reset()
			currentChunk.WriteString(overlapText)
			currentLength = len(overlapText)
			startOffset = startOffset + len(chunkText) - len(overlapText)
			chunkIndex++
		}
		
		currentChunk.WriteString(sentence)
		currentLength += sentenceLen
	}

	// Add remaining chunk
	if currentChunk.Len() > 0 {
		chunkText := currentChunk.String()
		chunks = append(chunks, Chunk{
			Text:           chunkText,
			Index:          chunkIndex,
			StartCharOffset: startOffset,
			EndCharOffset:  startOffset + len(chunkText),
		})
	}

	return chunks
}

// splitIntoSentences splits text into sentences
func (c *Chunker) splitIntoSentences(text string) []string {
	var sentences []string
	var currentSentence strings.Builder

	runes := []rune(text)
	
	for i, r := range runes {
		currentSentence.WriteRune(r)
		
		// Check for sentence endings
		if unicode.IsPunct(r) && (r == '.' || r == '!' || r == '?') {
			// Check if next character is space or end of text
			if i+1 >= len(runes) || unicode.IsSpace(runes[i+1]) {
				sentence := strings.TrimSpace(currentSentence.String())
				if len(sentence) > 0 {
					sentences = append(sentences, sentence+" ")
				}
				currentSentence.Reset()
			}
		}
	}

	// Add remaining text
	remaining := strings.TrimSpace(currentSentence.String())
	if len(remaining) > 0 {
		sentences = append(sentences, remaining+" ")
	}

	return sentences
}

// getOverlapText gets the last N characters from text for overlap
func (c *Chunker) getOverlapText(text string, overlapSize int) string {
	if len(text) <= overlapSize {
		return text
	}
	
	// Try to start overlap at sentence boundary
	overlapStart := len(text) - overlapSize
	for i := overlapStart; i < len(text); i++ {
		if i > 0 && (text[i] == '.' || text[i] == '!' || text[i] == '?') {
			if i+1 < len(text) && (text[i+1] == ' ' || text[i+1] == '\n') {
				return text[i+1:]
			}
		}
	}
	
	// Fallback: just take last N characters
	return text[len(text)-overlapSize:]
}

