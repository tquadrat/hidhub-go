package main

import (
	"flag"
	"fmt"

	"github.com/google/gousb"       // package gousb
	"github.com/google/gousb/usbid" // package usbid to identify a USB device
	"github.com/sstallion/go-hid"   // package hid
	"rsc.io/quote"
)

func main() {
	fmt.Println("Hello, World!")
	fmt.Println(quote.Go())

	//---* Define the command line options *-----------------------------------
	var vendorId int
	var productId int

	flag.IntVar(&vendorId, "vendorId", -1, "The vendor id of the keyboard to forward")
	flag.IntVar(&productId, "productId", -1, "The product id of the keyboard to forward")

	flag.Usage = func() {
		fmt.Println("Usage")
	}

	//---* Read the commandline *----------------------------------------------
	flag.Parse()

	if vendorId > 0 {
		fmt.Printf("VendorId  = 0x%04x\n", vendorId)
	}
	if productId > 0 {
		fmt.Printf("ProductId = 0x%04x\n", productId)
	}

	//---* Initalise GoUSB *---------------------------------------------------
	fmt.Println("Use the Package 'gousb'")

	//---* Create and initialise the Context *---------------------------------
	usbContext := gousb.NewContext()
	defer usbContext.Close()

	//devices,status :=
	usbContext.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		fmt.Printf("%03d.%03d %s:%s %s\n", desc.Bus, desc.Address, desc.Vendor, desc.Product, usbid.Describe(desc))
		fmt.Printf("ID %s:%s Class %s:%s Protocol %s\n",
			desc.Vendor.String(),
			desc.Product.String(),
			desc.Class.String(),
			desc.SubClass.String(),
			desc.Protocol.String())
		fmt.Printf("  Descriptionl: %s\n", usbid.Describe(desc))
		fmt.Printf("  Protocol    : %s\n", usbid.Classify(desc))

		// The configurations can be examined from the DeviceDesc, though they
		// can only be set once the device is opened. All configuration
		// references must be closed,  to free up the memory in libusb.
		for _, cfg := range desc.Configs {
			// This loop just uses more of the built-in and usbid pretty
			// printing to list/ the USB devices.
			fmt.Printf("  %s:\n", cfg)
			for _, intf := range cfg.Interfaces {
				fmt.Printf("    --------------\n")
				for _, ifSetting := range intf.AltSettings {
					fmt.Printf("    %s\n", ifSetting)
					fmt.Printf("      %s\n", usbid.Classify(ifSetting))
					fmt.Printf("      Class: %s SubClass: %s Protocol: %s\n", ifSetting.Class.String(), ifSetting.SubClass.String)
					for _, end := range ifSetting.Endpoints {
						fmt.Printf("      %s\n", end)
					}
				}
			}
			fmt.Printf("    --------------\n")
		}

		return false
	})

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
