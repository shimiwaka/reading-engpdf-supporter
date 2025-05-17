package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Page struct {
	Lines []string `json:"lines"`
}

type ExtractedData struct {
	Pages []Page `json:"pages"`
}

func extractTextFromPDF(pdfPath string) ([]string, error) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "pdftext_*")
	if err != nil {
		return nil, fmt.Errorf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 出力ファイルのパス
	outputPath := filepath.Join(tempDir, "output.txt")

	// pdftotextコマンドを実行
	cmd := exec.Command("pdftotext", "-layout", pdfPath, outputPath)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("テキスト抽出に失敗しました: %v", err)
	}

	// 抽出したテキストを読み込む
	content, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("テキストファイルの読み込みに失敗しました: %v", err)
	}

	// テキストを整形してページごとに分割
	text := string(content)
	text = strings.TrimSpace(text)
	pages := strings.Split(text, "\f")
	
	// 各ページの空行を削除
	var cleanedPages []string
	for _, page := range pages {
		page = strings.TrimSpace(page)
		if page != "" {
			cleanedPages = append(cleanedPages, page)
		}
	}

	return cleanedPages, nil
}

func isGoogleTranslation(text string) bool {
	return strings.Contains(text, "Machine Translated by Google")
}

func shouldIgnoreLine(text string) bool {
	return isGoogleTranslation(text) || strings.Contains(text, "©")
}

func splitIntoLines(text string) []string {
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	var currentLine strings.Builder

	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// 空行でない場合
		if line != "" && !shouldIgnoreLine(line) {
			// 現在の行に追加
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(line)

			// 次の行が空行または最後の行の場合、現在の行を保存
			if i == len(lines)-1 || strings.TrimSpace(lines[i+1]) == "" {
				cleanedLines = append(cleanedLines, currentLine.String())
				currentLine.Reset()
			}
		}
	}

	return cleanedLines
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("使用方法: go run main.go <英語PDFファイルのパス> <日本語PDFファイルのパス>")
		return
	}

	engPDFPath := os.Args[1]
	jpPDFPath := os.Args[2]

	// 英語PDFからテキストを抽出
	engPages, err := extractTextFromPDF(engPDFPath)
	if err != nil {
		fmt.Printf("英語PDFの処理に失敗しました: %v\n", err)
		return
	}

	// 日本語PDFからテキストを抽出
	jpPages, err := extractTextFromPDF(jpPDFPath)
	if err != nil {
		fmt.Printf("日本語PDFの処理に失敗しました: %v\n", err)
		return
	}

	// 英語データを準備
	engData := ExtractedData{
		Pages: make([]Page, len(engPages)),
	}
	for i, page := range engPages {
		engData.Pages[i] = Page{
			Lines: splitIntoLines(page),
		}
	}

	// 日本語データを準備
	jpData := ExtractedData{
		Pages: make([]Page, len(jpPages)),
	}
	for i, page := range jpPages {
		jpData.Pages[i] = Page{
			Lines: splitIntoLines(page),
		}
	}

	// JSONデータを文字列に変換
	engJSON, err := json.Marshal(engData)
	if err != nil {
		fmt.Printf("英語データのJSON変換に失敗しました: %v\n", err)
		return
	}
	jpJSON, err := json.Marshal(jpData)
	if err != nil {
		fmt.Printf("日本語データのJSON変換に失敗しました: %v\n", err)
		return
	}

	// テンプレートデータを準備
	type TemplateData struct {
		EnglishData  string
		JapaneseData string
	}
	templateData := TemplateData{
		EnglishData:  string(engJSON),
		JapaneseData: string(jpJSON),
	}

	// テンプレートを解析
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		fmt.Printf("テンプレートの解析に失敗しました: %v\n", err)
		return
	}

	// HTMLファイルを生成
	outputFile, err := os.Create("dialogue_comparison.html")
	if err != nil {
		fmt.Printf("出力ファイルの作成に失敗しました: %v\n", err)
		return
	}
	defer outputFile.Close()

	// テンプレートを実行
	if err := tmpl.Execute(outputFile, templateData); err != nil {
		fmt.Printf("HTMLの生成に失敗しました: %v\n", err)
		return
	}
} 