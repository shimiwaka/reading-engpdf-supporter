# English-Japanese Dialogue Comparison Tool

A tool that helps you compare English and Japanese dialogues from PDF files. This tool is particularly useful for language learners who want to study English with Japanese translations.

## Requirements

- Go 1.16 or later
- pdftotext (poppler-utils)

## Installation

1. Install poppler-utils:
   - Windows: Download from [poppler for Windows](http://blog.alivate.com.au/poppler-windows/)
   - Linux: `sudo apt-get install poppler-utils`
   - macOS: `brew install poppler`

2. Clone this repository:
   ```bash
   git clone https://github.com/yourusername/reading-engpdf-supporter.git
   cd reading-engpdf-supporter
   ```

## Usage

1. Run the program with English and Japanese PDF files:
   ```bash
   go run main.go english.pdf japanese.pdf
   ```

2. The program will generate `dialogue_comparison.html` in the current directory.

3. Open the HTML file in your web browser to view the comparison.

## How to Use

- Click on English text to toggle between English and Japanese
- Click the "次の日本語" (Next Japanese) button to add the next Japanese translation
   - When English text and Japanese text is unmatch, please use it
- Select English text to look up words in Japanese dictionaries
