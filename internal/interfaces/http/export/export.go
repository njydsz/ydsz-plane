// Package export 提供 HTTP 响应的表格导出工具（CSV / XLSX）。
//
// 设计目标：
//   - 与第三方库解耦：XLSX 使用纯 Go 标准库生成最小合法 OOXML 文件，
//     避免在服务端引入 excelize 等重量级依赖（对齐本项目"纯标准库"约定）。
//   - 统一响应头与文件名格式，避免各 handler 各自实现导致的风格漂移。
//   - CSV 输出携带 UTF-8 BOM，保证 Excel 直接打开中文不乱码。
//
// 用法示例：
//
//	export.WriteCSV(c, "defects.csv", headers, rows)
//	export.WriteXLSX(c, "缺陷明细", "defects.xlsx", headers, rows)
package export

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// WriteCSV 将表头与数据行以 CSV 格式写入响应。
// filename 仅用于 Content-Disposition，无需携带目录分隔符。
func WriteCSV(c *gin.Context, filename string, headers []string, rows [][]string) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// UTF-8 BOM：让 Excel 正确识别中文编码
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	if len(headers) > 0 {
		_ = w.Write(headers)
	}
	for _, row := range rows {
		_ = w.Write(row)
	}
	w.Flush()
}

// WriteXLSX 将表头与数据行以最小合法 .xlsx（OOXML）格式写入响应。
// sheetName 为工作表名；filename 用于 Content-Disposition。
func WriteXLSX(c *gin.Context, sheetName, filename string, headers []string, rows [][]string) {
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	// 辅助函数：向 ZIP 归档中添加文件
	addZIPFile := func(name string, content string) {
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		})
		if err != nil {
			return
		}
		fmt.Fprint(w, content)
	}

	// [Content_Types].xml — 声明部件内容类型
	addZIPFile("[Content_Types].xml",
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+
			`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`+
			`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`+
			`<Default Extension="xml" ContentType="application/xml"/>`+
			`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`+
			`<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`+
			`</Types>`)

	// _rels/.rels — 包级关系
	addZIPFile("_rels/.rels",
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+
			`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`+
			`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>`+
			`</Relationships>`)

	// xl/workbook.xml — 工作簿定义
	addZIPFile("xl/workbook.xml",
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+
			`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`+
			`<sheets><sheet name="`+xmlEscape(sheetName)+`" sheetId="1" r:id="rId1"/></sheets>`+
			`</workbook>`)

	// xl/_rels/workbook.xml.rels
	addZIPFile("xl/_rels/workbook.xml.rels",
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+
			`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`+
			`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>`+
			`</Relationships>`)

	// 组装表头 + 数据行
	var cells strings.Builder
	if len(headers) > 0 {
		cells.WriteString(buildXLSXRow(headers))
	}
	for _, row := range rows {
		cells.WriteString(buildXLSXRow(row))
	}

	// xl/worksheets/sheet1.xml — 数据表
	sheetXML := fmt.Sprintf(xlsxTemplate, cells.String())
	addZIPFile("xl/worksheets/sheet1.xml",
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+sheetXML)
}

// xlsxTemplate 是 OOXML 最小化模板，参考 ECMA-376 第 4 版 SpreadsheetML 规范。
const xlsxTemplate = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<cols>
  <col min="1" max="1" width="14" customWidth="1"/>
  <col min="2" max="2" width="10" customWidth="1"/>
  <col min="3" max="3" width="40" customWidth="1"/>
  <col min="4" max="4" width="12" customWidth="1"/>
  <col min="5" max="5" width="10" customWidth="1"/>
  <col min="6" max="6" width="10" customWidth="1"/>
  <col min="7" max="7" width="8" customWidth="1"/>
  <col min="8" max="8" width="16" customWidth="1"/>
  <col min="9" max="9" width="18" customWidth="1"/>
  <col min="10" max="10" width="18" customWidth="1"/>
</cols>
<sheetData>%s</sheetData></worksheet>`

// xlsxRow 是一行单元格 XML 片段。
const xlsxRow = `<row>%s</row>`

// xlsxCell 是一个单元格 XML 片段（内联字符串）。
const xlsxCell = `<c t="inlineStr"><is><t>%s</t></is></c>`

// buildXLSXRow 构建一行 xlsx 单元格 XML。
// 所有值作为内联字符串（t="inlineStr"）写入，避免数字/日期类型歧义。
func buildXLSXRow(vals []string) string {
	var cells strings.Builder
	for _, v := range vals {
		cells.WriteString(fmt.Sprintf(xlsxCell, xmlEscape(v)))
	}
	return fmt.Sprintf(xlsxRow, cells.String())
}

// xmlEscape 转义 XML 特殊字符（& < > " '）。
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// TimeNow 返回当前时间（可被测试覆盖）。
var TimeNow = func() time.Time { return time.Now() }
