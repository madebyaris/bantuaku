package regulations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/chat"
)

// KeywordGenerator generates UMKM-relevant search keywords using AI
type KeywordGenerator struct {
	chatProvider chat.ChatProvider
	chatModel    string
	log          logger.Logger
}

// NewKeywordGenerator creates a new keyword generator
func NewKeywordGenerator(chatProvider chat.ChatProvider, chatModel string) *KeywordGenerator {
	return &KeywordGenerator{
		chatProvider: chatProvider,
		chatModel:    chatModel,
		log:          *logger.Default(),
	}
}

// GenerateKeywords generates UMKM-relevant regulation search keywords in Bahasa Indonesia
func (kg *KeywordGenerator) GenerateKeywords(ctx context.Context) ([]string, error) {
	kg.log.Info("Generating UMKM regulation keywords with AI")

	prompt := `Kamu adalah ahli regulasi bisnis UMKM di Indonesia. Buatkan 25-30 keyword pencarian dalam Bahasa Indonesia untuk menemukan regulasi pemerintah yang relevan untuk UMKM.

KATEGORI YANG HARUS DICAKUP:
1. Perpajakan UMKM (PPh, PPN, NPWP)
2. Perizinan Usaha (NIB, SIUP, izin usaha)
3. Ketenagakerjaan (UMR, BPJS, kontrak kerja)
4. Keamanan Pangan (BPOM, sertifikasi halal, izin edar)
5. Hak Kekayaan Intelektual (merek dagang, paten UMKM)
6. Ekspor Impor (izin ekspor, bea cukai UMKM)
7. Lingkungan Hidup (AMDAL, izin lingkungan)
8. Standar Produk (SNI, sertifikasi)

ATURAN:
- Keyword dalam Bahasa Indonesia
- Sertakan istilah resmi (PP, Perpres, Permen, UU)
- Fokus pada regulasi yang berlaku untuk UMKM
- Sertakan variasi: "peraturan", "regulasi", "ketentuan", "persyaratan"

FORMAT OUTPUT:
Berikan JSON array of strings, contoh: ["keyword1", "keyword2", "keyword3"]
Hanya berikan JSON array tanpa penjelasan.`

	resp, err := kg.chatProvider.CreateChatCompletion(ctx, chat.ChatCompletionRequest{
		Model: kg.chatModel,
		Messages: []chat.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   800,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate keywords: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	// Parse JSON response
	var keywords []string
	content := resp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), &keywords); err != nil {
		kg.log.Warn("Failed to parse AI keywords response, using fallback", "error", err, "content", content[:min(200, len(content))])
		// Fallback keywords
		keywords = kg.getFallbackKeywords()
	}

	kg.log.Info("Generated keywords", "count", len(keywords))
	return keywords, nil
}

// getFallbackKeywords returns predefined keywords if AI generation fails
func (kg *KeywordGenerator) getFallbackKeywords() []string {
	return []string{
		// Perpajakan
		"peraturan pajak UMKM Indonesia",
		"PPh final UMKM 0.5 persen",
		"ketentuan PPN usaha kecil",
		"NPWP wajib usaha mikro",
		"PP 23 tahun 2018 pajak UMKM",
		// Perizinan
		"NIB nomor induk berusaha UMKM",
		"OSS perizinan usaha online",
		"izin usaha mikro kecil menengah",
		"persyaratan SIUP terbaru",
		"PP 5 tahun 2021 perizinan berusaha",
		// Ketenagakerjaan
		"UMR upah minimum regional terbaru",
		"BPJS ketenagakerjaan UMKM",
		"peraturan kontrak kerja karyawan",
		"PP ketenagakerjaan UMKM",
		"hak pekerja usaha kecil",
		// Keamanan Pangan
		"izin edar BPOM makanan minuman",
		"sertifikasi halal MUI UMKM",
		"persyaratan PIRT pangan",
		"standar keamanan pangan UMKM",
		"peraturan label kemasan makanan",
		// HAKI
		"pendaftaran merek dagang UMKM",
		"perlindungan HAKI usaha kecil",
		"paten sederhana UMKM",
		// Ekspor Impor
		"izin ekspor produk UMKM",
		"bea cukai barang ekspor",
		"persyaratan ekspor makanan",
		// Lingkungan & Standar
		"izin lingkungan usaha kecil",
		"SNI standar nasional Indonesia",
		"sertifikasi produk UMKM",
	}
}
