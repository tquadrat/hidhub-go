package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/sstallion/go-hid"   // package hid
)

func main() {
	//---* The usage text and other output *-----------------------------------
	usage := 
`
When running on a Raspberry Pi computer that is connected as a USB Gagdet 
to a host computer, this program forwards the input from the captured
keyboard to that host.

The captured keyboard is identified by its Vendor and its Product Id; if 
not known, it can be obtained through the tool 'lsusb'.
`
	          
	//---* Define the command line options *-----------------------------------
	var (
		vendorId int
		productId int
	)

	flag.IntVar( &vendorId, "vendorId", -1, "The vendor id of the keyboard to forward" )
	flag.IntVar( &productId, "productId", -1, "The product id of the keyboard to forward" )

	flag.Usage = func() {
		fmt.Fprintf( flag.CommandLine.Output(), "\nUsage of %s:\n\n", os.Args[0] )
	    flag.PrintDefaults()
	    fmt.Fprintln( flag.CommandLine.Output(), usage )
	}

	//---* Read the commandline *----------------------------------------------
	flag.Parse()

	if vendorId > 0 {
		fmt.Printf("VendorId  = 0x%04x\n", vendorId)
	}
	if productId > 0 {
		fmt.Printf("ProductId = 0x%04x\n", productId)
	}

	fmt.Println("------------------------------------------------")

	//---* Initialise HIDAPI *-------------------------------------------------
	fmt.Println("Use the Package 'hid'")
	fmt.Println("HIDAPI Version: ", hid.GetVersionStr())
	hid.Init()
	defer hid.Exit()

	//---* Dump the connected HID deviced *------------------------------------
	hid.Enumerate(hid.VendorIDAny, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		fmt.Printf("%s: ID %04x:%04x '%s' '%s'\n",
			info.Path,
			info.VendorID,
			info.ProductID,
			info.MfrStr,
			info.ProductStr)
		return nil
	})
} //  main()