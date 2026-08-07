package export

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// newTestCtx 构造带响应记录器的 gin.Context。
func newTestCtx() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestWriteCSV(t *testing.T) {
	c, w := newTestCtx()
	headers := []string{"编号", "名称", "严重程度"}
	rows := [][]string{
		{"YD-1", "登录页崩溃", "S1(致命)"},
		{"YD-2", "导出乱码", "S3(一般)"},
	}

	WriteCSV(c, "defects-export.csv", headers, rows)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.Contains(cd, "defects-export.csv") {
		t.Errorf("Content-Disposition = %q, want filename=defects-export.csv", cd)
	}

	body := w.Body.Bytes()
	// UTF-8 BOM 必须存在（Excel 兼容）
	if len(body) < 3 || body[0] != 0xEF || body[1] != 0xBB || body[2] != 0xBF {
		t.Error("CSV 输出缺少 UTF-8 BOM")
	}

	// 解析 CSV 验证内容
	r := csv.NewReader(bytes.NewReader(body[3:]))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("CSV 解析失败: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("records = %d 行, want 3（1 表头 + 2 数据）", len(records))
	}
	if records[0][0] != "编号" || records[1][0] != "YD-1" || records[2][2] != "S3(一般)" {
		t.Errorf("CSV 内容不正确: %v", records)
	}
}

func TestWriteXLSX_ValidZIP(t *testing.T) {
	c, w := newTestCtx()
	headers := []string{"编号", "名称"}
	rows := [][]string{{"YD-1", "含特殊字符 & < > \" '"}}

	WriteXLSX(c, "缺陷明细", "defects-export.xlsx", headers, rows)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "openxmlformats") {
		t.Errorf("Content-Type = %q, want openxmlformats", ct)
	}

	// 必须是合法的 ZIP 归档
	zr, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if err != nil {
		t.Fatalf("输出不是合法 ZIP（xlsx 结构）: %v", err)
	}

	found := map[string]bool{}
	for _, f := range zr.File {
		found[f.Name] = true
	}
	for _, required := range []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"xl/workbook.xml",
		"xl/_rels/workbook.xml.rels",
		"xl/worksheets/sheet1.xml",
	} {
		if !found[required] {
			t.Errorf("xlsx 缺少必需部件: %s", required)
		}
	}

	// 校验 sheet1.xml 中特殊字符被转义
	sheet := readZIPFile(t, zr, "xl/worksheets/sheet1.xml")
	for _, want := range []string{"&amp;", "&lt;", "&gt;", "&quot;", "&apos;"} {
		if !strings.Contains(sheet, want) {
			t.Errorf("sheet1.xml 未转义 XML 特殊字符，缺少 %s", want)
		}
	}
}

func TestBuildXLSXRow_XML_Escape(t *testing.T) {
	row := buildXLSXRow([]string{`a&b<c>"d"'e'`})
	if !strings.Contains(row, "&amp;") || !strings.Contains(row, "&lt;") ||
		!strings.Contains(row, "&gt;") || !strings.Contains(row, "&quot;") || !strings.Contains(row, "&apos;") {
		t.Errorf("buildXLSXRow 未完整转义: %s", row)
	}
}

func readZIPFile(t *testing.T, zr *zip.Reader, name string) string {
	t.Helper()
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("打开 %s: %v", name, err)
			}
			defer rc.Close()
			b, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("读取 %s: %v", name, err)
			}
			return string(b)
		}
	}
	t.Fatalf("ZIP 中不存在 %s", name)
	return ""
}
