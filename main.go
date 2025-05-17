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

const htmlTemplate = `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>英語・日本語対照表示</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        .dialogue-container {
            display: flex;
            margin-bottom: 20px;
            gap: 20px;
        }
        .english {
            flex: 1;
            padding: 10px;
            border: 1px solid #ccc;
            border-radius: 5px;
            cursor: pointer;
        }
        .japanese {
            flex: 1;
            padding: 10px;
            border: 1px solid #ccc;
            border-radius: 5px;
            white-space: pre-wrap;
            cursor: pointer;
        }
        .page-break {
            border-top: 2px dashed #999;
            margin: 20px 0;
        }
        button {
            padding: 5px 10px;
            margin-left: 10px;
            cursor: pointer;
        }
        .hidden {
            display: none;
        }
    </style>
</head>
<body>
    <div id="content"></div>

    <script>
        const englishData = {{.EnglishData}};
        const japaneseData = {{.JapaneseData}};
        let currentLineIndices = {};

        function displayAllData() {
            const content = document.getElementById('content');
            content.innerHTML = '';

            englishData.pages.forEach((engPage, pageIndex) => {
                engPage.lines.forEach((engLine, lineIndex) => {
                    const container = document.createElement('div');
                    container.className = 'dialogue-container';

                    const engDiv = document.createElement('div');
                    engDiv.className = 'english';
                    engDiv.textContent = engLine;
                    engDiv.onclick = () => toggleDisplay(pageIndex, lineIndex);

                    const jpDiv = document.createElement('div');
                    jpDiv.className = 'japanese hidden';
                    const key = pageIndex + '_' + lineIndex;
                    jpDiv.textContent = japaneseData.pages[pageIndex].lines[lineIndex] || '';
                    jpDiv.onclick = () => toggleDisplay(pageIndex, lineIndex);

                    const button = document.createElement('button');
                    button.textContent = '次の日本語';
                    button.className = 'hidden';
                    button.onclick = () => {
                        if (!currentLineIndices[key]) {
                            currentLineIndices[key] = lineIndex;
                        }
                        currentLineIndices[key]++;

                        // 現在の行の日本語テキストを更新
                        const nextIndex = currentLineIndices[key];
                        const nextText = japaneseData.pages[pageIndex].lines[nextIndex] || '';
                        if (nextText) {
                            jpDiv.textContent = jpDiv.textContent + '\n' + nextText;
                        }

                        // 次の行以降の日本語テキストを更新
                        for (let i = lineIndex + 1; i < engPage.lines.length; i++) {
                            const nextKey = pageIndex + '_' + i;
                            if (!currentLineIndices[nextKey]) {
                                currentLineIndices[nextKey] = i;
                            }
                            currentLineIndices[nextKey]++;
                            const nextDiv = document.querySelector('[data-key="' + nextKey + '"]');
                            if (nextDiv) {
                                const nextLineIndex = currentLineIndices[nextKey];
                                nextDiv.textContent = japaneseData.pages[pageIndex].lines[nextLineIndex] || '';
                            }
                        }
                    };

                    jpDiv.setAttribute('data-key', key);
                    container.appendChild(engDiv);
                    container.appendChild(jpDiv);
                    container.appendChild(button);
                    content.appendChild(container);
                });

                // ページ区切りを表示（最後のページ以外）
                if (pageIndex < englishData.pages.length - 1) {
                    const pageBreak = document.createElement('div');
                    pageBreak.className = 'page-break';
                    content.appendChild(pageBreak);
                }
            });
        }

        function toggleDisplay(pageIndex, lineIndex) {
            const key = pageIndex + '_' + lineIndex;
            const engDiv = document.querySelector('[data-key="' + key + '"]').previousElementSibling;
            const jpDiv = document.querySelector('[data-key="' + key + '"]');
            const button = jpDiv.nextElementSibling;

            if (jpDiv.classList.contains('hidden')) {
                // 英語から英語+日本語に切り替え
                jpDiv.classList.remove('hidden');
                button.classList.remove('hidden');
            } else {
                // 英語+日本語から英語のみに切り替え
                jpDiv.classList.add('hidden');
                button.classList.add('hidden');
            }
        }

        // データを表示
        displayAllData();
    </script>
</body>
</html>`

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
	tmpl, err := template.New("html").Parse(htmlTemplate)
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