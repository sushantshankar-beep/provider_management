package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"provider_management/internal/domain"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

func (h *SettlementHandler) exportCSV(c *gin.Context, settlements  []domain.ProviderSettlement) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	headers := []string{
		"Provider ID",
		"Provider Name",
		"Account No",
		"IFSC Code",
		"Net Amount",
		"Narration",
		"Status",
	}
	writer.Write(headers)

	for _, s := range settlements {
		row := []string{
			s.ProviderID.Hex(),
			s.ProviderName,
			s.AccountNo,
			s.IfscCode,
			fmt.Sprintf("%.2f", s.TotalAmount),
			string(s.Status),
		}
		writer.Write(row)
	}

	writer.Flush()

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=settlements_%s.csv", time.Now().Format("20060102_150405")))
	c.Data(200, "text/csv", buf.Bytes())
}

func (h *SettlementHandler) exportExcel(c *gin.Context, settlements  []domain.ProviderSettlement) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Settlements"
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)

	headers := []string{
		"Provider ID",
		"Provider Name",
		"Account No",
		"IFSC Code",
		"Net Amount",
		"Status",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheet, cell, header)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheet, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), headerStyle)

	for i, s := range settlements {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), s.ProviderID.Hex())
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), s.ProviderName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), s.AccountNo)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), s.IfscCode)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), s.TotalAmount)
		// f.SetCellValue(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf(s.PayoutIdNumber))
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), s.Status)
	}

	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c", 'A'+i)
		f.SetColWidth(sheet, col, col, 15)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		c.JSON(500, gin.H{"error": true, "message": err.Error()})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=settlements_%s.xlsx", time.Now().Format("20060102_150405")))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *SettlementHandler) exportPDF(c *gin.Context, settlements  []domain.ProviderSettlement) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Settlement Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(68, 114, 196)
	pdf.SetTextColor(255, 255, 255)

	headers := []string{"Provider", "Account No", "IFSC", "Net Amount", "Status"}
	widths := []float64{45, 35, 30, 25, 30, 25}

	for i, header := range headers {
		pdf.CellFormat(widths[i], 7, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(240, 240, 240)

	fill := false
	for _, s := range settlements {
		pdf.CellFormat(widths[0], 6, s.ProviderName, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(widths[1], 6, s.AccountNo, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(widths[2], 6, s.IfscCode, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(widths[3], 6, fmt.Sprintf("%.2f", math.Round(s.TotalAmount*100)/100) , "1", 0, "R", fill, 0, "")
		// pdf.CellFormat(widths[4], 6, fmt.Sprintf(s.PayoutIdNumber), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(widths[5], 6, string(s.Status), "1", 0, "C", fill, 0, "")
		pdf.Ln(-1)
		fill = !fill
	}

	buf := new(bytes.Buffer)
	err := pdf.Output(buf)
	if err != nil {
		c.JSON(500, gin.H{"error": true, "message": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=settlements_%s.pdf", time.Now().Format("20060102_150405")))
	c.Data(200, "application/pdf", buf.Bytes())
}