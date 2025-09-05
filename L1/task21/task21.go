package main

import (
	"fmt"
	"os"
)

type PrinterV1 interface {
	Print(msg string) error
}

type PrinterV2 interface {
	PrintStored()
}

type LegacyPrinter struct{}

func (l *LegacyPrinter) Print(msg string) error {
	_, err := fmt.Fprintln(os.Stdout, "LegacyPrinter:", msg)
	return err
}

type ModernPrinter struct {
	msg string
}

func (m *ModernPrinter) PrintStored() {
	fmt.Fprintln(os.Stdout, "ModernPrinter:", m.msg)
}

type PrinterAdapter struct {
	oldPrinter PrinterV1
	msg        string
}

func NewPrinterAdapter (printer PrinterV1, msg string) *PrinterAdapter {
	return &PrinterAdapter{
		oldPrinter: printer,
		msg: msg,
	}
}

func (a *PrinterAdapter) PrintStored() {
	newMsg := "ModernPrinter: " + a.msg
	err := a.oldPrinter.Print(newMsg)
	if err != nil {
		a.oldPrinter.Print(err.Error())
	}
}

func main() {
	oldPrinter := LegacyPrinter{}
	message := "Message to print"
	adapter := NewPrinterAdapter(&oldPrinter, message)
	
	adapter.PrintStored()
}
